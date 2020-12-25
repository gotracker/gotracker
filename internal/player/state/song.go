package state

import (
	"fmt"
	"time"

	"github.com/gotracker/gomixing/mixing"
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"
	device "github.com/gotracker/gosound"

	"gotracker/internal/player/feature"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/render"
	"gotracker/internal/player/state/pattern"
)

// Song is the song state for the current playing song
type Song struct {
	intf.Song
	SongData           intf.SongData
	EffectFactory      intf.EffectFactoryFunc
	CalcSemitonePeriod intf.CalcSemitonePeriodFunc

	Channels     []ChannelState
	Pattern      pattern.State
	GlobalVolume volume.Volume

	rowTxn intf.SongPositionState
}

// NewSong creates a new song structure and sets its default values
func NewSong() *Song {
	var ss = Song{}
	ss.Pattern.Reset()

	return &ss
}

// GetNumChannels returns the number of channels
func (ss *Song) GetNumChannels() int {
	return len(ss.Channels)
}

// SetNumChannels updates the song to have the specified number of channels and resets their states
func (ss *Song) SetNumChannels(num int) {
	ss.Channels = make([]ChannelState, num)

	for ch, cs := range ss.Channels {
		cs.Pos = sampling.Pos{}
		cs.PrevInstrument = nil
		cs.Instrument = nil
		cs.Period = 0
		cs.Command = nil

		cs.DisplayNote = note.EmptyNote
		cs.DisplayInst = 0

		cs.TargetPeriod = cs.Period
		cs.TargetPos = cs.Pos
		cs.TargetInst = nil
		cs.PortaTargetPeriod = cs.TargetPeriod
		cs.NotePlayTick = 0
		cs.RetriggerCount = 0
		cs.TremorOn = true
		cs.TremorTime = 0
		cs.VibratoDelta = 0
		cs.Cmd = nil
		cs.OutputChannelNum = ss.SongData.GetOutputChannel(ch)
	}
}

// RenderOneRow renders the next single row from the song pattern data into a RowRender object
func (ss *Song) RenderOneRow(sampler *render.Sampler) (*device.PremixData, error) {
	// pre-mix row updates
	{
		patIdx, err := ss.Pattern.GetCurrentPatternIdx()
		if err != nil {
			return nil, err
		}

		pat := ss.SongData.GetPattern(patIdx)
		if pat == nil {
			return nil, pattern.ErrStopSong
		}

		rows := pat.GetRows()
		rowTxn := ss.Pattern.StartTransaction()
		defer func() {
			rowTxn.Cancel()
			ss.rowTxn = nil
		}()
		ss.rowTxn = rowTxn

		myCurrentRow := ss.Pattern.GetCurrentRow()

		row := rows[myCurrentRow]
		for channelNum, channel := range row.GetChannels() {
			if channelNum >= ss.GetNumChannels() {
				continue
			}

			cs := &ss.Channels[channelNum]

			cs.processRow(row, channel, ss, ss.SongData, ss.EffectFactory, ss.CalcSemitonePeriod, ss.processCommand)
		}

		rowTxn.Commit()
	}

	finalData := &render.RowRender{}
	premix := &device.PremixData{
		Userdata: finalData,
	}

	// row render
	{
		rowTxn := ss.Pattern.StartTransaction()
		defer func() {
			rowTxn.Cancel()
			ss.rowTxn = nil
		}()
		ss.rowTxn = rowTxn

		ss.soundRenderRow(premix, sampler)
		nCh := 0
		for ch := range ss.Channels {
			if !ss.SongData.IsChannelEnabled(ch) {
				continue
			}
			nCh++
		}
		var rowText = render.NewRowText(nCh)
		for ch := range ss.Channels {
			if !ss.SongData.IsChannelEnabled(ch) {
				continue
			}
			cs := &ss.Channels[ch]
			c := render.ChannelDisplay{
				Note:       cs.DisplayNote.String(),
				Instrument: "..",
				Volume:     "..",
				Effect:     "...",
			}

			if cs.DisplayInst != 0 {
				c.Instrument = fmt.Sprintf("%0.2d", cs.DisplayInst)
			}

			if cs.DisplayVolume != volume.VolumeUseInstVol {
				c.Volume = fmt.Sprintf("%0.2d", uint8(cs.DisplayVolume*64.0))
			}

			if cs.Cmd != nil {
				if cs.ActiveEffect != nil {
					c.Effect = cs.ActiveEffect.String()
				}
			}
			rowText[ch] = c
		}
		finalData.Order = int(ss.Pattern.GetCurrentOrder())
		finalData.Row = int(ss.Pattern.GetCurrentRow())
		finalData.RowText = rowText

		rowTxn.AdvanceRow()

		rowTxn.Commit()
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

	if cs.TargetPeriod == 0 && cs.Instrument != nil && cs.Instrument.GetKeyOn() {
		cs.Instrument.SetKeyOn(cs.PrevNoteSemitone, false)
	} else if cs.DoRetriggerNote && currentTick == cs.NotePlayTick {
		cs.Instrument = nil
		if cs.TargetInst != nil {
			if cs.PrevInstrument != nil && cs.PrevInstrument.GetInstrument() == cs.TargetInst {
				cs.Instrument = cs.PrevInstrument
				cs.Instrument.SetKeyOn(cs.PrevNoteSemitone, false)
			} else {
				cs.Instrument = cs.TargetInst.InstantiateOnChannel(cs.OutputChannelNum, cs.Filter)
			}
		}
		cs.Period = cs.TargetPeriod
		cs.Pos = cs.TargetPos
		if cs.Period != 0 && cs.Instrument != nil {
			cs.Instrument.SetKeyOn(cs.NoteSemitone, true)
		}
	}
}

func (ss *Song) soundRenderRow(premix *device.PremixData, sampler *render.Sampler) {
	mix := sampler.Mixer()

	samplerSpeed := sampler.GetSamplerSpeed()
	tickDuration := time.Duration(2500) * time.Millisecond / time.Duration(ss.Pattern.GetTempo())
	samplesPerTick := int(tickDuration.Seconds() * float64(sampler.SampleRate))

	ticksThisRow := ss.Pattern.GetTicksThisRow()

	samplesThisRow := int(ticksThisRow) * samplesPerTick

	panmixer := sampler.GetPanMixer()

	centerPanning := panmixer.GetMixingMatrix(panning.CenterAhead)

	for len(premix.Data) < len(ss.Channels) {
		premix.Data = append(premix.Data, nil)
	}
	premix.SamplesLen = samplesThisRow

	for ch := range ss.Channels {
		cs := &ss.Channels[ch]
		if ss.SongData.IsChannelEnabled(ch) {
			rr := make([]mixing.Data, ticksThisRow)
			cs.renderRow(rr, ch, ticksThisRow, mix, panmixer, samplerSpeed, samplesPerTick, centerPanning, tickDuration)

			premix.Data[ch] = rr
		}
	}
}

// SetNextOrder sets the next order index
func (ss *Song) SetNextOrder(order intf.OrderIdx) {
	if ss.rowTxn != nil {
		ss.rowTxn.SetNextOrder(order)
	} else {
		rowTxn := ss.Pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetNextOrder(order)
		rowTxn.Commit()
	}
}

// SetNextRow sets the next row index
func (ss *Song) SetNextRow(row intf.RowIdx) {
	if ss.rowTxn != nil {
		ss.rowTxn.SetNextRow(row)
	} else {
		rowTxn := ss.Pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetNextRow(row)
		rowTxn.Commit()
	}
}

// SetTempo sets the desired tempo for the song
func (ss *Song) SetTempo(tempo int) {
	if ss.rowTxn != nil {
		ss.rowTxn.SetTempo(tempo)
	} else {
		rowTxn := ss.Pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetTempo(tempo)
		rowTxn.Commit()
	}
}

// DecreaseTempo reduces the tempo by the `delta` value
func (ss *Song) DecreaseTempo(delta int) {
	if ss.rowTxn != nil {
		ss.rowTxn.AccTempoDelta(-delta)
	} else {
		rowTxn := ss.Pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.AccTempoDelta(-delta)
		rowTxn.Commit()
	}
}

// IncreaseTempo increases the tempo by the `delta` value
func (ss *Song) IncreaseTempo(delta int) {
	if ss.rowTxn != nil {
		ss.rowTxn.AccTempoDelta(delta)
	} else {
		rowTxn := ss.Pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.AccTempoDelta(delta)
		rowTxn.Commit()
	}
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
	if ss.rowTxn != nil {
		ss.rowTxn.SetTicks(ticks)
	} else {
		rowTxn := ss.Pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetTicks(ticks)
		rowTxn.Commit()
	}
}

// AddRowTicks increases the number of ticks the row expects to play for
func (ss *Song) AddRowTicks(ticks int) {
	if ss.rowTxn != nil {
		ss.rowTxn.SetFinePatternDelay(ticks)
	} else {
		rowTxn := ss.Pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetFinePatternDelay(ticks)
		rowTxn.Commit()
	}
}

// SetPatternDelay sets the repeat number for the row to `rept`
// NOTE: this may be set 1 time (first in wins) and will be reset only by the next row being read in
func (ss *Song) SetPatternDelay(rept int) {
	if ss.rowTxn != nil {
		ss.rowTxn.SetPatternDelay(rept)
	} else {
		rowTxn := ss.Pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetPatternDelay(rept)
		rowTxn.Commit()
	}
}

// SetPatternLoopStart sets the pattern loop start position
func (ss *Song) SetPatternLoopStart() {
	if ss.rowTxn != nil {
		ss.rowTxn.SetPatternLoopStart()
	} else {
		rowTxn := ss.Pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetPatternLoopStart()
		rowTxn.Commit()
	}
}

// SetPatternLoopEnd sets the pattern loop end position
func (ss *Song) SetPatternLoopEnd() {
	if ss.rowTxn != nil {
		ss.rowTxn.SetPatternLoopEnd()
	} else {
		rowTxn := ss.Pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetPatternLoopEnd()
		rowTxn.Commit()
	}
}

// SetPatternLoopCount sets the total loops desired for the pattern loop mechanism
func (ss *Song) SetPatternLoopCount(loops int) {
	if ss.rowTxn != nil {
		ss.rowTxn.SetPatternLoopCount(loops)
	} else {
		rowTxn := ss.Pattern.StartTransaction()
		defer rowTxn.Cancel()

		rowTxn.SetPatternLoopCount(loops)
		rowTxn.Commit()
	}
}

// DisableFeatures disables specified features
func (ss *Song) DisableFeatures(features []feature.Feature) {
	for _, f := range features {
		switch f {
		case feature.PatternLoop:
			ss.Pattern.PatternLoopEnabled = false
		}
	}
}

// CanPatternLoop returns true if the song is allowed to pattern loop
func (ss *Song) CanPatternLoop() bool {
	return ss.Pattern.PatternLoopEnabled
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
func (ss *Song) SetOrderList(orders []intf.PatternIdx) {
	ss.Pattern.Orders = orders
}

// SetSongData sets the song data object
func (ss *Song) SetSongData(songdata intf.SongData) {
	ss.SongData = songdata
}

// GetSongData gets the song data object
func (ss *Song) GetSongData() intf.SongData {
	return ss.SongData
}

// GetChannel returns the channel interface for the specified channel number
func (ss *Song) GetChannel(ch int) intf.Channel {
	return &ss.Channels[ch]
}

// GetCurrentOrder returns the current order
func (ss *Song) GetCurrentOrder() intf.OrderIdx {
	return ss.Pattern.GetCurrentOrder()
}

// GetCurrentRow returns the current row
func (ss *Song) GetCurrentRow() intf.RowIdx {
	return ss.Pattern.GetCurrentRow()
}
