package state

import (
	"errors"
	"fmt"

	"github.com/heucuva/gomixing/mixing"
	"github.com/heucuva/gomixing/panning"
	"github.com/heucuva/gomixing/sampling"
	"github.com/heucuva/gomixing/volume"
	device "github.com/heucuva/gosound"

	"gotracker/internal/player/feature"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/render"
)

var (
	// ErrStopSong is a magic error asking to stop the current song
	ErrStopSong = errors.New("stop song")
)

// Song is the song state for the current playing song
type Song struct {
	intf.Song
	SongData           intf.SongData
	EffectFactory      intf.EffectFactoryFunc
	CalcSemitonePeriod intf.CalcSemitonePeriodFunc

	Channels     []ChannelState
	Pattern      PatternState
	GlobalVolume volume.Volume

	PatternLoopEnabled bool
	playedOrders       []uint8 // when PatternLoopEnabled is false, this is used to detect loops
}

// NewSong creates a new song structure and sets its default values
func NewSong() *Song {
	var ss = Song{}
	ss.Pattern.CurrentOrder = 0
	ss.Pattern.CurrentRow = 0
	ss.PatternLoopEnabled = true
	ss.playedOrders = make([]uint8, 0)

	return &ss
}

// GetNumChannels returns the number of channels
func (ss *Song) GetNumChannels() int {
	return len(ss.Channels)
}

// SetNumChannels updates the song to have the specified number of channels and resets their states
func (ss *Song) SetNumChannels(num int) {
	ss.Channels = make([]ChannelState, num)

	for _, cs := range ss.Channels {
		cs.Pos = sampling.Pos{}
		cs.Instrument = nil
		cs.Period = 0
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
}

// RenderOneRow renders the next single row from the song pattern data into a RowRender object
func (ss *Song) RenderOneRow(sampler *render.Sampler) (*device.PremixData, error) {
	ol := ss.SongData.GetOrderList()
	if ss.Pattern.CurrentOrder < 0 || int(ss.Pattern.CurrentOrder) >= len(ol) {
		return nil, ErrStopSong
	}
	patNum := PatternNum(ol[ss.Pattern.CurrentOrder])
	if patNum == NextPattern {
		ss.Pattern.CurrentOrder++
		return nil, nil
	}

	if patNum == InvalidPattern {
		ss.Pattern.CurrentOrder++
		return nil, nil // this is supposed to be a song break
	}

	pattern := ss.SongData.GetPattern(uint8(patNum))
	if pattern == nil {
		return nil, ErrStopSong
	}

	rows := pattern.GetRows()
	if ss.Pattern.CurrentRow < 0 || int(ss.Pattern.CurrentRow) >= len(rows) {
		ss.Pattern.CurrentRow = 0
		ss.Pattern.CurrentOrder++
		return nil, nil
	}

	orderRestart := false
	rowRestart := false

	ss.Pattern.RowHasPatternDelay = false
	ss.Pattern.PatternDelay = 0
	ss.Pattern.FinePatternDelay = 0

	finalData := &render.RowRender{}
	premix := &device.PremixData{
		Userdata: finalData,
	}

	if int(ss.Pattern.CurrentRow) >= len(rows) {
		orderRestart = true
		ss.Pattern.CurrentOrder++
	} else {
		//myCurrentOrder := ss.Pattern.CurrentOrder
		myCurrentRow := ss.Pattern.CurrentRow

		row := rows[myCurrentRow]
		for channelNum, channel := range row.GetChannels() {
			if !ss.SongData.IsChannelEnabled(channelNum) {
				continue
			}

			cs := &ss.Channels[channelNum]

			isOr, isRr := cs.processRow(row, channel, ss, ss.SongData, ss.EffectFactory, ss.CalcSemitonePeriod, ss.processCommand)
			orderRestart = orderRestart || isOr
			rowRestart = rowRestart || isRr
		}

		ss.soundRenderRow(premix, sampler)
		var rowText = render.NewRowText(len(ss.Channels))
		for ch := range ss.Channels {
			cs := &ss.Channels[ch]
			c := render.ChannelDisplay{
				Note:       "...",
				Instrument: "..",
				Volume:     "..",
				Effect:     "...",
			}

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

	currentOrder := ss.Pattern.CurrentOrder
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

	if !ss.PatternLoopEnabled && currentOrder != ss.Pattern.CurrentOrder {
		ss.playedOrders = append(ss.playedOrders, ss.Pattern.CurrentOrder)
	}

	return premix, nil
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

func (ss *Song) soundRenderRow(premix *device.PremixData, sampler *render.Sampler) {
	mix := sampler.Mixer()

	samplerSpeed := sampler.GetSamplerSpeed()
	tickSamples := int(2.5 * float32(sampler.SampleRate) / float32(ss.Pattern.Row.Tempo))

	rowLoops := 1
	if ss.Pattern.RowHasPatternDelay {
		rowLoops = ss.Pattern.PatternDelay
	}
	extraTicks := ss.Pattern.FinePatternDelay

	ticksThisRow := int(ss.Pattern.Row.Ticks)*rowLoops + extraTicks

	samplesThisRow := int(ticksThisRow) * tickSamples

	panmixer := sampler.GetPanMixer()

	centerPanning := panmixer.GetMixingMatrix(panning.CenterAhead)

	for len(premix.Data) < len(ss.Channels) {
		premix.Data = append(premix.Data, nil)
	}
	premix.SamplesLen = samplesThisRow

	for ch := range ss.Channels {
		cs := &ss.Channels[ch]
		rr := make([]mixing.Data, ticksThisRow)
		cs.renderRow(rr, ch, ticksThisRow, mix, panmixer, samplerSpeed, tickSamples, centerPanning)

		premix.Data[ch] = rr
	}
}

// SetCurrentOrder sets the current order index
func (ss *Song) SetCurrentOrder(order uint8) {
	if !ss.PatternLoopEnabled {
		for _, o := range ss.playedOrders {
			if o == order {
				return
			}
		}
	}
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

// GetGlobalVolume returns the global volume value
func (ss *Song) GetGlobalVolume() volume.Volume {
	return ss.GlobalVolume
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

// DisableFeatures disables specified features
func (ss *Song) DisableFeatures(features []feature.Feature) {
	for _, f := range features {
		switch f {
		case feature.PatternLoop:
			ss.PatternLoopEnabled = false
		}
	}
}

// CanPatternLoop returns true if the song is allowed to pattern loop
func (ss *Song) CanPatternLoop() bool {
	return ss.PatternLoopEnabled
}

// SetEffectFactory sets the active effect factory function
func (ss *Song) SetEffectFactory(ef intf.EffectFactoryFunc) {
	ss.EffectFactory = ef
}

// SetCalcSemitonePeriod sets the semitone period calculator function
func (ss *Song) SetCalcSemitonePeriod(csp intf.CalcSemitonePeriodFunc) {
	ss.CalcSemitonePeriod = csp
}

// SetPatterns sets the pattern list interface
func (ss *Song) SetPatterns(patterns intf.Patterns) {
	ss.Pattern.Patterns = patterns
}

// SetOrderList sets the order list
func (ss *Song) SetOrderList(orders []uint8) {
	ss.Pattern.Orders = orders
}

// SetSongData sets the song data object
func (ss *Song) SetSongData(songdata intf.SongData) {
	ss.SongData = songdata
}

// GetChannel returns the channel interface for the specified channel number
func (ss *Song) GetChannel(ch int) intf.Channel {
	return &ss.Channels[ch]
}

// GetCurrentOrder returns the current order
func (ss *Song) GetCurrentOrder() uint8 {
	return ss.Pattern.CurrentOrder
}

// GetCurrentRow returns the current row
func (ss *Song) GetCurrentRow() uint8 {
	return ss.Pattern.CurrentRow
}
