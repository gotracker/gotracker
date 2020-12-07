package state

import (
	"fmt"
	"gotracker/internal/player/channel"
	"gotracker/internal/player/instrument"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/render"
	"gotracker/internal/s3m"
	s3mEffect "gotracker/internal/s3m/effect"
	"gotracker/internal/s3m/util"
	"gotracker/internal/s3m/volume"
	"log"
	"math"
)

type EffectFactory func(mi intf.SharedMemory, data channel.Data) intf.Effect

type Song struct {
	intf.Song
	SongData      *s3m.S3M
	EffectFactory EffectFactory

	Channels     [32]ChannelState
	NumChannels  int
	Pattern      PatternState
	SampleMult   volume.Volume
	GlobalVolume volume.Volume
}

func CreateSongState(filename string) *Song {
	song, err := s3m.ReadS3M(filename)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	var ss = &Song{}
	ss.EffectFactory = s3mEffect.EffectFactory
	ss.Pattern.Patterns = &song.Patterns
	ss.Pattern.Orders = &song.Head.OrderList
	ss.Pattern.Row.Ticks = int(song.Head.Info.InitialSpeed)
	ss.Pattern.Row.Tempo = int(song.Head.Info.InitialTempo)
	ss.Pattern.CurrentOrder = 0
	ss.Pattern.CurrentRow = 0
	ss.SampleMult = 1.0
	ss.GlobalVolume = volume.FromS3M(song.Head.Info.GlobalVolume)
	ss.SongData = song
	ss.NumChannels = 1

	// old method for determining active channels
	// for _, pattern := range *ss.Pattern.Patterns {
	// 	if ss.NumChannels == 32 {
	// 		break
	// 	}
	// 	for _, row := range pattern.Rows {
	// 		if ss.NumChannels == 32 {
	// 			break
	// 		}
	// 		for i := ss.NumChannels; i < 32; i++ {
	// 			channel := row[i]
	// 			if channel.What.HasCommand() || channel.What.HasNote() {
	// 				ss.NumChannels = i + 1
	// 			}
	// 		}
	// 	}
	// }

	// new method for determining active channels (uses S3M data I somehow overlooked before)
	for i, cs := range ss.SongData.Head.ChannelSettings {
		if cs.IsEnabled() {
			ss.NumChannels = i + 1
		}
	}

	for i := 0; i < ss.NumChannels; i++ {
		cs := &ss.Channels[i]
		cs.Instrument = instrument.InstrumentInfo{}
		cs.Pos = 0
		cs.Period = 0
		cs.SetStoredVolume(64, ss)
		ch := song.Head.ChannelSettings[i]
		if ch.IsEnabled() {
			pf := song.Head.Panning[i]
			if pf.IsValid() {
				cs.Pan = pf.Value()
			} else {
				l := ch.GetChannel()
				switch l {
				case s3m.ChannelIDL1, s3m.ChannelIDL2, s3m.ChannelIDL3, s3m.ChannelIDL4, s3m.ChannelIDL5, s3m.ChannelIDL6, s3m.ChannelIDL7, s3m.ChannelIDL8:
					cs.Pan = 0x03
				case s3m.ChannelIDR1, s3m.ChannelIDR2, s3m.ChannelIDR3, s3m.ChannelIDR4, s3m.ChannelIDR5, s3m.ChannelIDR6, s3m.ChannelIDR7, s3m.ChannelIDR8:
					cs.Pan = 0x0C
				}
			}
		} else {
			cs.Pan = 0x08 // center?
		}
		cs.Command = nil

		cs.DisplayNote = note.EmptyNote
		cs.DisplayInst = 0

		cs.TargetPeriod = cs.Period
		cs.TargetPos = cs.Pos
		cs.TargetInst = cs.Instrument
		cs.PortaTargetPeriod = cs.TargetPeriod
		cs.NotePlayTick = 0
		cs.RetriggerCount = 0
		cs.TremorOn = true
		cs.TremorTime = 0
		cs.VibratoDelta = 0
		cs.Cmd = nil
	}
	return ss
}

func (ss *Song) RenderNextRow(sampler *render.Sampler) []byte {
	var pattern = ss.Pattern.GetRow()
	if pattern == nil {
		ss.Pattern.NextRow()
		return nil
	}

	bSetOrder := false
	nextOrder := uint8(0)
	bSetRow := false
	nextRow := uint8(0)

	if bSetOrder || bSetRow {
		if bSetOrder {
			ss.Pattern.CurrentOrder = nextOrder
		}
		if bSetRow {
			ss.Pattern.CurrentRow = nextRow
		}
	} else {
		ss.Pattern.NextRow()
	}

	return []byte{}
}

func (ss *Song) RenderOneRow(sampler *render.Sampler) *render.RowRender {
	if ss.Pattern.CurrentOrder < 0 || int(ss.Pattern.CurrentOrder) >= len(ss.SongData.Head.OrderList) {
		var done = &render.RowRender{}
		done.Stop = true
		return done
	}
	patNum := PatternNum(ss.SongData.Head.OrderList[ss.Pattern.CurrentOrder])
	if patNum == NextPattern {
		ss.Pattern.CurrentOrder++
		return nil
	}

	if patNum == InvalidPattern {
		ss.Pattern.CurrentOrder++
		return nil // this is supposed to be a song break
	}

	pattern := ss.SongData.Patterns[patNum]
	if &pattern == nil {
		var done = &render.RowRender{}
		done.Stop = true
		return done
	}

	if ss.Pattern.CurrentRow < 0 || int(ss.Pattern.CurrentRow) >= len(pattern.Rows) {
		ss.Pattern.CurrentRow = 0
		ss.Pattern.CurrentOrder++
		return nil
	}

	var orderRestart = false
	var rowRestart = false

	ss.Pattern.RowHasPatternDelay = false
	ss.Pattern.PatternDelay = 0
	ss.Pattern.FinePatternDelay = 0

	var finalData = &render.RowRender{}
	finalData.Stop = false

	if int(ss.Pattern.CurrentRow) > len(pattern.Rows) {
		orderRestart = true
		ss.Pattern.CurrentOrder++
	} else {
		myCurrentOrder := ss.Pattern.CurrentOrder
		myCurrentRow := ss.Pattern.CurrentRow

		row := &pattern.Rows[myCurrentRow]
		for channelNum, channel := range row {
			if !ss.SongData.Head.ChannelSettings[channelNum].IsEnabled() {
				continue
			}

			cs := &ss.Channels[channelNum]

			cs.Command = nil

			cs.TargetPeriod = cs.Period
			cs.TargetPos = cs.Pos
			cs.TargetInst = cs.Instrument
			cs.PortaTargetPeriod = cs.TargetPeriod
			cs.NotePlayTick = 0
			cs.RetriggerCount = 0
			cs.TremorOn = true
			cs.TremorTime = 0
			cs.VibratoDelta = 0
			cs.Cmd = &row[channelNum]

			wantNoteCalc := false

			if channel.What.HasNote() {
				cs.VibratoOscillator.Pos = 0
				cs.TremoloOscillator.Pos = 0
				cs.TargetInst = instrument.InstrumentInfo{}
				if channel.Instrument == 0 {
					// use current
					cs.TargetInst = cs.Instrument
					cs.TargetPos = 0
				} else if int(channel.Instrument) > len(ss.SongData.Instruments) {
					cs.TargetInst = instrument.InstrumentInfo{}
				} else {
					cs.TargetInst.Sample = &ss.SongData.Instruments[channel.Instrument-1]
					cs.TargetInst.Id = channel.Instrument
					cs.TargetPos = 0
					if !cs.TargetInst.IsInvalid() {
						cs.SetStoredVolume(cs.TargetInst.Sample.Info.Volume, ss)
					}
				}

				if channel.Note.IsInvalid() {
					cs.TargetPeriod = 0
					cs.DisplayNote = note.EmptyNote
					cs.DisplayInst = 0
				} else if !cs.TargetInst.IsInvalid() {
					cs.NoteSemitone = channel.Note.Semitone()
					cs.TargetC2Spd = cs.TargetInst.C2Spd()
					wantNoteCalc = true
					cs.DisplayNote = channel.Note
					cs.DisplayInst = cs.TargetInst.Id
				}
			} else {
				cs.DisplayNote = note.EmptyNote
				cs.DisplayInst = 0
			}

			if channel.What.HasVolume() {
				if channel.Volume == 255 {
					if !cs.Instrument.IsInvalid() {
						cs.SetStoredVolume(cs.Instrument.Sample.Info.Volume, ss)
					}
				} else {
					cs.SetStoredVolume(channel.Volume, ss)
				}
			}

			cs.ActiveEffect = ss.EffectFactory(cs, *cs.Cmd)

			if wantNoteCalc {
				cs.TargetPeriod = util.CalcSemitonePeriod(cs.NoteSemitone, cs.TargetC2Spd)
			}

			if cs.Cmd.What.HasCommand() {
				cs.SetEffectSharedMemoryIfNonZero(cs.Cmd.Info)
			}
			if cs.ActiveEffect != nil {
				cs.ActiveEffect.PreStart(cs, ss)
			}
			if ss.Pattern.CurrentOrder != myCurrentOrder {
				orderRestart = true
			}
			if ss.Pattern.CurrentRow != myCurrentRow {
				rowRestart = true
			}

			cs.Command = ss.processCommand
		}

		ss.soundRenderRow(finalData, sampler)
		var rowText = render.NewRowText(ss.NumChannels)
		for ch := 0; ch < ss.NumChannels; ch++ {
			cs := &ss.Channels[ch]

			var c render.ChannelDisplay
			c.Note = "..."
			c.Instrument = ".."
			c.Volume = ".."
			c.Effect = "..."

			if !cs.Instrument.IsInvalid() && cs.Period != 0 {
				c.Note = cs.DisplayNote.String()
			}

			if cs.DisplayInst != 0 {
				c.Instrument = fmt.Sprintf("%0.2d", cs.DisplayInst)
			}

			if cs.Cmd != nil {
				if cs.Cmd.What.HasVolume() {
					c.Volume = fmt.Sprintf("%0.2d", cs.Cmd.Volume)
				}

				if cs.Cmd.What.HasCommand() {
					c.Effect = fmt.Sprintf("%c%0.2x", cs.Cmd.Command+'@', cs.Cmd.Info)
				}
			}
			rowText[ch] = c
		}
		finalData.Order = int(ss.Pattern.CurrentOrder)
		finalData.Row = int(ss.Pattern.CurrentRow)
		finalData.RowText = rowText
	}

	if !rowRestart {
		if orderRestart {
			ss.Pattern.CurrentRow = 0
		} else {
			if ss.Pattern.LoopEnabled {
				if ss.Pattern.CurrentRow == ss.Pattern.LoopEnd {
					ss.Pattern.LoopCount++
					if ss.Pattern.LoopCount >= ss.Pattern.LoopTotal {
						ss.Pattern.CurrentRow++
						ss.Pattern.LoopEnabled = false
					} else {
						ss.Pattern.CurrentRow = ss.Pattern.LoopStart
					}
				} else {
					ss.Pattern.CurrentRow++
				}
			} else {
				ss.Pattern.CurrentRow++
			}
		}
	} else if !orderRestart {
		ss.Pattern.CurrentOrder++
	}

	if ss.Pattern.CurrentRow >= 64 {
		ss.Pattern.CurrentRow = 0
		ss.Pattern.CurrentOrder++
	}

	return finalData
}

func (ss *Song) processCommand(ch int, cs *ChannelState, currentTick int, lastTick bool) {
	if cs.ActiveEffect != nil {
		if currentTick == 0 {
			cs.ActiveEffect.Start(cs, ss)
		}
		cs.ActiveEffect.Tick(cs, ss, currentTick)
		if lastTick {
			cs.ActiveEffect.Stop(cs, ss, currentTick)
		}
	}

	if currentTick == cs.NotePlayTick {
		cs.Instrument = cs.TargetInst
		cs.Period = cs.TargetPeriod
		cs.Pos = cs.TargetPos
	}
}

func (ss *Song) soundRenderRow(rowRender *render.RowRender, sampler *render.Sampler) {
	samplerSpeed := sampler.GetSamplerSpeed()
	tickSamples := 2.5 * float32(sampler.SampleRate) / float32(ss.Pattern.Row.Tempo)

	rowLoops := 1
	if ss.Pattern.RowHasPatternDelay {
		rowLoops = ss.Pattern.PatternDelay
	}
	extraTicks := ss.Pattern.FinePatternDelay

	ticksThisRow := int(ss.Pattern.Row.Ticks)*rowLoops + extraTicks

	samples := int(tickSamples * float32(ticksThisRow))

	data := make([]volume.Volume, sampler.Channels*samples)

	for ch := 0; ch < ss.NumChannels; ch++ {
		cs := &ss.Channels[ch]

		tickPos := 0
		for tick := 0; tick < ticksThisRow; tick++ {
			simulatedTick := tick % ss.Pattern.Row.Ticks
			var lastTick = (tick+1 == ticksThisRow)
			if cs.Command != nil {
				cs.Command(ch, cs, simulatedTick, lastTick)
			}
			if !cs.Instrument.IsInvalid() && cs.Period != 0 {
				period := cs.Period + cs.VibratoDelta
				samplerAdd := samplerSpeed / period

				vol := volume.FromS3M(cs.ActiveVolume) * cs.LastGlobalVolume
				pan := volume.Volume(cs.Pan) / 16.0
				volL := vol * (1.0 - pan)
				volR := vol * pan

				for s := 0; s < int(tickSamples); s++ {
					if !cs.PlaybackFrozen() {
						if (cs.Instrument.Sample.Info.Flags & 1) != 0 {
							if int(cs.Pos) >= int(cs.Instrument.Sample.Info.LoopEndL) {
								cs.Pos = float32(cs.Instrument.Sample.Info.LoopBeginL)
							}
							if cs.Pos < 0 {
								cs.Pos = 0
							}
						}
						if int(cs.Pos) < len(cs.Instrument.Sample.Sample) {
							samp := (volume.Volume(cs.Instrument.Sample.Sample[int(cs.Pos)]) - 128.0) / 128.0
							if sampler.Channels == 1 {
								data[tickPos] += samp * vol
							} else {
								data[tickPos] += samp * volL
								data[tickPos+1] += samp * volR
							}
						}
						cs.Pos += samplerAdd
						if cs.Pos < 0 {
							cs.Pos = 0
						}
					}
					tickPos += sampler.Channels
				}
			}
		}
	}

	ss.SampleMult = 1.0
	for _, sample := range data {
		ss.SampleMult = volume.Volume(math.Max(float64(ss.SampleMult), math.Abs(float64(sample))))
	}

	rowRender.RenderData = make([]byte, sampler.Channels*(sampler.BitsPerSample/8)*samples)
	oidx := 0
	sampleDivisor := 1.0 / ss.SampleMult
	for _, sample := range data {
		sample *= sampleDivisor
		if sampler.BitsPerSample == 8 {
			rowRender.RenderData[oidx] = sample.ToByte()
			oidx++
		} else {
			val := int16(sample * 16384.0)
			rowRender.RenderData[oidx] = byte(val & 0xFF)
			rowRender.RenderData[oidx+1] = byte(val >> 8)
			oidx += 2
		}
	}
}

func (ss *Song) SetCurrentOrder(order uint8) {
	ss.Pattern.CurrentOrder = order
}

func (ss *Song) SetCurrentRow(row uint8) {
	ss.Pattern.CurrentRow = row
}

func (ss *Song) SetTempo(tempo int) {
	ss.Pattern.Row.Tempo = tempo
}

func (ss *Song) DecreaseTempo(delta int) {
	ss.Pattern.Row.Tempo -= delta
}

func (ss *Song) IncreaseTempo(delta int) {
	ss.Pattern.Row.Tempo += delta
}

func (ss *Song) SetGlobalVolume(vol volume.Volume) {
	ss.GlobalVolume = vol
}

func (ss *Song) SetTicks(ticks int) {
	ss.Pattern.Row.Ticks = ticks
}

func (ss *Song) AddRowTicks(ticks int) {
	ss.Pattern.FinePatternDelay += ticks
}

func (ss *Song) SetPatternDelay(rept int) {
	if !ss.Pattern.RowHasPatternDelay {
		ss.Pattern.RowHasPatternDelay = true
		ss.Pattern.PatternDelay = rept
	}
}

func (ss *Song) SetPatternLoopStart() {
	ss.Pattern.LoopStart = ss.Pattern.CurrentRow
}

func (ss *Song) SetPatternLoopEnd(loops uint8) {
	ss.Pattern.LoopEnd = ss.Pattern.CurrentRow
	ss.Pattern.LoopTotal = loops
	if !ss.Pattern.LoopEnabled {
		ss.Pattern.LoopEnabled = true
		ss.Pattern.LoopCount = 0
	}
}
