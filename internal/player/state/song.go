package state

import (
	"fmt"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/render"
	"gotracker/internal/player/volume"
)

// EffectFactory is a function that generates a channel effect based on the input channel pattern data
type EffectFactory func(mi intf.SharedMemory, data intf.ChannelData) intf.Effect

// SemitoneCalculator is the function used to calculate a note semitone
type SemitoneCalculator func(noteSemi note.Semitone, c2spd note.C2SPD) note.Period

// Song is the song state for the current playing song
type Song struct {
	intf.Song
	SongData           intf.SongData
	EffectFactory      EffectFactory
	CalcSemitonePeriod SemitoneCalculator

	Channels     [32]ChannelState
	NumChannels  int
	Pattern      PatternState
	SampleMult   volume.Volume
	GlobalVolume volume.Volume
}

// NewSong creates a new song structure and sets its default values
func NewSong() *Song {
	var ss = Song{}
	ss.Pattern.CurrentOrder = 0
	ss.Pattern.CurrentRow = 0
	ss.SampleMult = 1.0
	ss.NumChannels = 1

	return &ss
}

// RenderOneRow renders the next single row from the song pattern data into a RowRender object
func (ss *Song) RenderOneRow(sampler *render.Sampler) *render.RowRender {
	ol := ss.SongData.GetOrderList()
	if ss.Pattern.CurrentOrder < 0 || int(ss.Pattern.CurrentOrder) >= len(ol) {
		var done = &render.RowRender{}
		done.Stop = true
		return done
	}
	patNum := PatternNum(ol[ss.Pattern.CurrentOrder])
	if patNum == NextPattern {
		ss.Pattern.CurrentOrder++
		return nil
	}

	if patNum == InvalidPattern {
		ss.Pattern.CurrentOrder++
		return nil // this is supposed to be a song break
	}

	pattern := ss.SongData.GetPattern(uint8(patNum))
	if pattern == nil {
		var done = &render.RowRender{}
		done.Stop = true
		return done
	}

	rows := pattern.GetRows()
	if ss.Pattern.CurrentRow < 0 || int(ss.Pattern.CurrentRow) >= len(rows) {
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

	if int(ss.Pattern.CurrentRow) > len(rows) {
		orderRestart = true
		ss.Pattern.CurrentOrder++
	} else {
		myCurrentOrder := ss.Pattern.CurrentOrder
		myCurrentRow := ss.Pattern.CurrentRow

		row := rows[myCurrentRow]
		for channelNum, channel := range row.GetChannels() {
			if !ss.SongData.IsChannelEnabled(channelNum) {
				continue
			}

			cs := &ss.Channels[channelNum]

			cs.Command = nil

			cs.TargetPeriod = cs.Period
			cs.TargetPos = cs.Pos
			cs.TargetInst = cs.Instrument
			cs.DoRetriggerNote = true
			cs.NotePlayTick = 0
			cs.RetriggerCount = 0
			cs.TremorOn = true
			cs.TremorTime = 0
			cs.VibratoDelta = 0
			cs.Cmd = channel

			wantNoteCalc := false

			if channel.HasNote() {
				cs.VibratoOscillator.Pos = 0
				cs.TremoloOscillator.Pos = 0
				cs.TargetInst = nil
				inst := channel.GetInstrument()
				if inst == 0 {
					// use current
					cs.TargetInst = cs.Instrument
					cs.TargetPos = 0
				} else if int(inst) > ss.SongData.NumInstruments() {
					cs.TargetInst = nil
				} else {
					cs.TargetInst = ss.SongData.GetInstrument(int(inst) - 1)
					cs.TargetPos = 0
					if cs.TargetInst != nil {
						vol := cs.TargetInst.GetVolume()
						cs.SetStoredVolume(vol, ss)
					}
				}

				n := channel.GetNote()
				if n.IsInvalid() {
					cs.TargetPeriod = 0
					cs.DisplayNote = note.EmptyNote
					cs.DisplayInst = 0
				} else if cs.TargetInst != nil {
					cs.NoteSemitone = n.Semitone()
					cs.TargetC2Spd = cs.TargetInst.GetC2Spd()
					wantNoteCalc = true
					cs.DisplayNote = n
					cs.DisplayInst = uint8(cs.TargetInst.GetID())
				}
			} else {
				cs.DisplayNote = note.EmptyNote
				cs.DisplayInst = 0
			}

			if channel.HasVolume() {
				v := channel.GetVolume()
				if v == volume.VolumeUseInstVol {
					sample := cs.TargetInst
					if sample != nil {
						vol := sample.GetVolume()
						cs.SetStoredVolume(vol, ss)
					}
				} else {
					cs.SetStoredVolume(v, ss)
				}
			}

			cs.ActiveEffect = ss.EffectFactory(cs, cs.Cmd)

			if wantNoteCalc {
				cs.TargetPeriod = ss.CalcSemitonePeriod(cs.NoteSemitone, cs.TargetC2Spd)
				cs.PortaTargetPeriod = cs.TargetPeriod
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

			if cs.TargetInst != nil && cs.Period != 0 {
				c.Note = cs.DisplayNote.String()
			}

			if cs.DisplayInst != 0 {
				c.Instrument = fmt.Sprintf("%0.2d", cs.DisplayInst)
			}

			if cs.Cmd != nil {
				if cs.Cmd.HasVolume() {
					c.Volume = fmt.Sprintf("%0.2d", uint8(cs.Cmd.GetVolume()*64.0))
				}

				if cs.ActiveEffect != nil {
					c.Effect = cs.ActiveEffect.String()
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

	if cs.DoRetriggerNote && currentTick == cs.NotePlayTick {
		cs.Instrument = cs.TargetInst
		cs.Period = cs.TargetPeriod
		cs.Pos = cs.TargetPos
	}
}

func (ss *Song) soundRenderRow(rowRender *render.RowRender, sampler *render.Sampler) {
	mixer := sampler.Mixer()

	samplerSpeed := sampler.GetSamplerSpeed()
	tickSamples := 2.5 * float32(sampler.SampleRate) / float32(ss.Pattern.Row.Tempo)

	rowLoops := 1
	if ss.Pattern.RowHasPatternDelay {
		rowLoops = ss.Pattern.PatternDelay
	}
	extraTicks := ss.Pattern.FinePatternDelay

	ticksThisRow := int(ss.Pattern.Row.Ticks)*rowLoops + extraTicks

	samples := int(tickSamples * float32(ticksThisRow))

	panmixer := sampler.GetPanMixer()

	data := mixer.NewMixBuffer(sampler.Channels, samples)

	for ch := 0; ch < ss.NumChannels; ch++ {
		cs := &ss.Channels[ch]

		tickPos := 0
		for tick := 0; tick < ticksThisRow; tick++ {
			var lastTick = (tick+1 == ticksThisRow)
			if cs.Command != nil {
				cs.Command(ch, cs, tick, lastTick)
			}

			sample := cs.Instrument
			if sample != nil && cs.Period != 0 {
				period := cs.Period + cs.VibratoDelta
				samplerAdd := samplerSpeed / float32(period)

				vol := cs.ActiveVolume * cs.LastGlobalVolume

				panmix := panmixer.GetMixingMatrix(float32(cs.Pan) / 16.0)

				for s := 0; s < int(tickSamples); s++ {
					if !cs.PlaybackFrozen() {
						if cs.Pos < 0 {
							cs.Pos = 0
						}
						samp := sample.GetSample(cs.Pos) * vol
						for c, pm := range panmix {
							data[c][tickPos] += samp * pm
						}
						cs.Pos += samplerAdd
					}
					tickPos++
				}
			}
		}
	}

	rowRender.RenderData = sampler.ToRenderData(data, ss.NumChannels)
}

// SetCurrentOrder sets the current order index
func (ss *Song) SetCurrentOrder(order uint8) {
	ss.Pattern.CurrentOrder = order
}

// SetCurrentRow sets the current row index
func (ss *Song) SetCurrentRow(row uint8) {
	ss.Pattern.CurrentRow = row
}

// SetTempo sets the desired tempo for the song
func (ss *Song) SetTempo(tempo int) {
	ss.Pattern.Row.Tempo = tempo
}

// DecreaseTempo reduces the tempo by the `delta` value
func (ss *Song) DecreaseTempo(delta int) {
	ss.Pattern.Row.Tempo -= delta
}

// IncreaseTempo increases the tempo by the `delta` value
func (ss *Song) IncreaseTempo(delta int) {
	ss.Pattern.Row.Tempo += delta
}

// SetGlobalVolume sets the global volume to the specified `vol` value
func (ss *Song) SetGlobalVolume(vol volume.Volume) {
	ss.GlobalVolume = vol
}

// SetTicks sets the number of ticks the row expects to play for
func (ss *Song) SetTicks(ticks int) {
	ss.Pattern.Row.Ticks = ticks
}

// AddRowTicks increases the number of ticks the row expects to play for
func (ss *Song) AddRowTicks(ticks int) {
	ss.Pattern.FinePatternDelay += ticks
}

// SetPatternDelay sets the repeat number for the row to `rept`
// NOTE: this may be set 1 time (first in wins) and will be reset only by the next row being read in
func (ss *Song) SetPatternDelay(rept int) {
	if !ss.Pattern.RowHasPatternDelay {
		ss.Pattern.RowHasPatternDelay = true
		ss.Pattern.PatternDelay = rept
	}
}

// SetPatternLoopStart sets the pattern loop start position
func (ss *Song) SetPatternLoopStart() {
	ss.Pattern.LoopStart = ss.Pattern.CurrentRow
}

// SetPatternLoopEnd sets the loop end position (and total loops desired)
func (ss *Song) SetPatternLoopEnd(loops uint8) {
	ss.Pattern.LoopEnd = ss.Pattern.CurrentRow
	ss.Pattern.LoopTotal = loops
	if !ss.Pattern.LoopEnabled {
		ss.Pattern.LoopEnabled = true
		ss.Pattern.LoopCount = 0
	}
}
