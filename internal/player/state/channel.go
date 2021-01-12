package state

import (
	"time"

	"github.com/gotracker/gomixing/mixing"
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

type commandFunc func(int, *ChannelState, int, bool)

// ChannelState is the state of a single channel
type ChannelState struct {
	intf.Channel

	activeState ActiveState
	targetState RenderState
	prevState   RenderState

	ActiveEffect intf.Effect

	TargetSemitone note.Semitone // from pattern, modified

	StoredSemitone    note.Semitone // from pattern, unmodified, current note
	DoRetriggerNote   bool
	PortaTargetPeriod note.Period
	NotePlayTick      int
	RetriggerCount    uint8
	Memory            intf.Memory
	TrackData         intf.ChannelData
	freezePlayback    bool
	LastGlobalVolume  volume.Volume
	Semitone          note.Semitone // from TargetSemitone, modified further, used in period calculations
	WantNoteCalc      bool
	WantVolCalc       bool
	UseTargetPeriod   bool
	volumeActive      bool
	PanEnabled        bool

	OutputChannelNum int
	Filter           intf.Filter
}

// Process processes a channel's row data
func (cs *ChannelState) Process(row intf.Row, globalVol volume.Volume, sd intf.SongData) {
	cs.prevState = cs.activeState.RenderState
	cs.targetState = cs.activeState.RenderState
	cs.DoRetriggerNote = true
	cs.NotePlayTick = 0
	cs.RetriggerCount = 0
	cs.activeState.PeriodDelta = 0

	cs.WantNoteCalc = false
	cs.WantVolCalc = false
	cs.UseTargetPeriod = false

	if cs.TrackData == nil {
		return
	}

	if cs.TrackData.HasNote() {
		cs.UseTargetPeriod = true
		inst := cs.TrackData.GetInstrument()
		if inst.IsEmpty() {
			// use current
			cs.targetState.Pos = sampling.Pos{}
		} else if !sd.IsValidInstrumentID(inst) {
			cs.targetState.Instrument = nil
		} else {
			cs.targetState.Instrument = sd.GetInstrument(inst)
			cs.targetState.Pos = sampling.Pos{}
			if cs.targetState.Instrument != nil {
				cs.WantVolCalc = true
			}
		}

		n := cs.TrackData.GetNote()
		if n == note.EmptyNote {
			cs.WantNoteCalc = false
			cs.DoRetriggerNote = cs.TrackData.HasInstrument()
			if cs.DoRetriggerNote {
				cs.targetState.Pos = sampling.Pos{}
			}
		} else if n.IsInvalid() {
			cs.targetState.Period = nil
			cs.WantNoteCalc = false
			cs.DoRetriggerNote = false
		} else if n == note.StopNote {
			cs.targetState.Period = cs.activeState.Period
			if cs.prevState.Instrument != nil {
				cs.targetState.Instrument = cs.prevState.Instrument
			}
			cs.WantNoteCalc = false
			cs.DoRetriggerNote = false
		} else if cs.targetState.Instrument != nil {
			cs.StoredSemitone = n.Semitone()
			cs.TargetSemitone = cs.StoredSemitone
			cs.WantNoteCalc = true
		}
	} else {
		cs.WantNoteCalc = false
		cs.WantVolCalc = false
		cs.DoRetriggerNote = false
	}

	if cs.TrackData.HasVolume() {
		cs.WantVolCalc = false
		v := cs.TrackData.GetVolume()
		if v == volume.VolumeUseInstVol {
			if cs.targetState.Instrument != nil {
				cs.WantVolCalc = true
			}
		} else {
			cs.SetActiveVolume(v)
		}
	}
}

// RenderRowTick renders a channel's row data for a single tick
func (cs *ChannelState) RenderRowTick(mix *mixing.Mixer, panmixer mixing.PanMixer, samplerSpeed float32, tickSamples int, tickDuration time.Duration) (*mixing.Data, error) {
	if cs.PlaybackFrozen() {
		return nil, nil
	}

	return cs.activeState.Render(cs.LastGlobalVolume, mix, panmixer, samplerSpeed, tickSamples, tickDuration)
}

// ResetStates resets the channel's internal states
func (cs *ChannelState) ResetStates() {
	cs.activeState.Reset()
	cs.targetState.Reset()
	cs.prevState.Reset()
}

// FreezePlayback suspends mixer progression on the channel
func (cs *ChannelState) FreezePlayback() {
	cs.freezePlayback = true
}

// UnfreezePlayback resumes mixer progression on the channel
func (cs *ChannelState) UnfreezePlayback() {
	cs.freezePlayback = false
}

// PlaybackFrozen returns true if the mixer progression for the channel is suspended
func (cs ChannelState) PlaybackFrozen() bool {
	return cs.freezePlayback
}

// ResetRetriggerCount sets the retrigger count to 0
func (cs *ChannelState) ResetRetriggerCount() {
	cs.RetriggerCount = 0
}

// GetMemory returns the interface to the custom effect memory module
func (cs *ChannelState) GetMemory() intf.Memory {
	return cs.Memory
}

// SetMemory sets the custom effect memory interface
func (cs *ChannelState) SetMemory(mem intf.Memory) {
	cs.Memory = mem
}

// GetActiveVolume returns the current active volume on the channel
func (cs *ChannelState) GetActiveVolume() volume.Volume {
	return cs.activeState.Volume
}

// SetActiveVolume sets the active volume on the channel
func (cs *ChannelState) SetActiveVolume(vol volume.Volume) {
	if vol != volume.VolumeUseInstVol {
		cs.activeState.Volume = vol
	}
}

// GetData returns the interface to the current channel song pattern data
func (cs *ChannelState) GetData() intf.ChannelData {
	return cs.TrackData
}

// GetPortaTargetPeriod returns the current target portamento (to note) sampler period
func (cs *ChannelState) GetPortaTargetPeriod() note.Period {
	return cs.PortaTargetPeriod
}

// SetPortaTargetPeriod sets the current target portamento (to note) sampler period
func (cs *ChannelState) SetPortaTargetPeriod(period note.Period) {
	cs.PortaTargetPeriod = period
}

// GetTargetPeriod returns the soon-to-be-committed sampler period (when the note retriggers)
func (cs *ChannelState) GetTargetPeriod() note.Period {
	return cs.targetState.Period
}

// SetTargetPeriod sets the soon-to-be-committed sampler period (when the note retriggers)
func (cs *ChannelState) SetTargetPeriod(period note.Period) {
	cs.targetState.Period = period
}

// SetVibratoDelta sets the vibrato (ephemeral) delta sampler period
func (cs *ChannelState) SetVibratoDelta(delta note.PeriodDelta) {
	cs.activeState.PeriodDelta = delta
}

// GetVibratoDelta gets the vibrato (ephemeral) delta sampler period
func (cs *ChannelState) GetVibratoDelta() note.PeriodDelta {
	return cs.activeState.PeriodDelta
}

// SetVolumeActive enables or disables the sample of the instrument
func (cs *ChannelState) SetVolumeActive(on bool) {
	cs.volumeActive = on
}

// GetInstrument returns the interface to the active instrument
func (cs *ChannelState) GetInstrument() intf.Instrument {
	return cs.activeState.Instrument
}

// SetInstrument sets the interface to the active instrument
func (cs *ChannelState) SetInstrument(inst intf.Instrument, pb intf.Playback) {
	cs.activeState.Instrument = inst
	if inst != nil {
		cs.activeState.NoteControl = inst.InstantiateOnChannel(cs.OutputChannelNum, cs.Filter)
		cs.activeState.NoteControl.SetPlayback(pb)
	}
}

// GetNoteControl returns the active note-control interface
func (cs *ChannelState) GetNoteControl() intf.NoteControl {
	return cs.activeState.NoteControl
}

// GetTargetInst returns the interface to the soon-to-be-committed active instrument (when the note retriggers)
func (cs *ChannelState) GetTargetInst() intf.Instrument {
	return cs.targetState.Instrument
}

// SetTargetInst sets the soon-to-be-committed active instrument (when the note retriggers)
func (cs *ChannelState) SetTargetInst(inst intf.Instrument) {
	cs.targetState.Instrument = inst
}

// GetNoteSemitone returns the note semitone for the channel
func (cs *ChannelState) GetNoteSemitone() note.Semitone {
	return cs.StoredSemitone
}

// GetTargetPos returns the soon-to-be-committed sample position of the instrument
func (cs *ChannelState) GetTargetPos() sampling.Pos {
	return cs.targetState.Pos
}

// SetTargetPos sets the soon-to-be-committed sample position of the instrument
func (cs *ChannelState) SetTargetPos(pos sampling.Pos) {
	cs.targetState.Pos = pos
}

// GetPeriod returns the current sampler period of the active instrument
func (cs *ChannelState) GetPeriod() note.Period {
	return cs.activeState.Period
}

// SetPeriod sets the current sampler period of the active instrument
func (cs *ChannelState) SetPeriod(period note.Period) {
	cs.activeState.Period = period
}

// GetPos returns the sample position of the active instrument
func (cs *ChannelState) GetPos() sampling.Pos {
	return cs.activeState.Pos
}

// SetPos sets the sample position of the active instrument
func (cs *ChannelState) SetPos(pos sampling.Pos) {
	cs.activeState.Pos = pos
}

// SetNotePlayTick sets the tick on which the note will retrigger
func (cs *ChannelState) SetNotePlayTick(tick int) {
	cs.NotePlayTick = tick
}

// GetRetriggerCount returns the current count of the retrigger counter
func (cs *ChannelState) GetRetriggerCount() uint8 {
	return cs.RetriggerCount
}

// SetRetriggerCount sets the current count of the retrigger counter
func (cs *ChannelState) SetRetriggerCount(cnt uint8) {
	cs.RetriggerCount = cnt
}

// SetPanEnabled activates or deactivates the panning. If enabled, then pan updates work (see SetPan)
func (cs *ChannelState) SetPanEnabled(on bool) {
	cs.PanEnabled = on
}

// SetPan sets the active panning value of the channel
func (cs *ChannelState) SetPan(pan panning.Position) {
	if cs.PanEnabled {
		cs.activeState.Pan = pan
	}
}

// GetPan gets the active panning value of the channel
func (cs *ChannelState) GetPan() panning.Position {
	return cs.activeState.Pan
}

// SetDoRetriggerNote sets the enablement flag for DoRetriggerNote
// this gets reset on every row
func (cs *ChannelState) SetDoRetriggerNote(enabled bool) {
	cs.DoRetriggerNote = enabled
	cs.UseTargetPeriod = enabled
	if enabled {
		cs.WantNoteCalc = true
	}
}

// SetSemitone sets the target semitone for the channel
func (cs *ChannelState) SetSemitone(st note.Semitone) {
	cs.TargetSemitone = st
}

// GetFilter returns the active filter on the channel
func (cs *ChannelState) GetFilter() intf.Filter {
	return cs.Filter
}

// SetFilter sets the active filter on the channel
func (cs *ChannelState) SetFilter(filter intf.Filter) {
	cs.Filter = filter
	if cs.activeState.NoteControl != nil {
		cs.activeState.NoteControl.SetFilter(filter)
	}
}

// SetOutputChannelNum sets the output channel number for the channel
func (cs *ChannelState) SetOutputChannelNum(outputChNum int) {
	cs.OutputChannelNum = outputChNum
}

// SetGlobalVolume sets the last-known global volume on the channel
func (cs *ChannelState) SetGlobalVolume(gv volume.Volume) {
	cs.LastGlobalVolume = gv
}
