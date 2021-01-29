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
	activeState activeState
	targetState intf.PlaybackState
	prevState   activeState

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
	Semitone          note.Semitone // from TargetSemitone, modified further, used in period calculations
	WantNoteCalc      bool
	WantVolCalc       bool
	UseTargetPeriod   bool
	volumeActive      bool
	PanEnabled        bool
	NewNoteAction     note.Action
	pastNote          []*activeState

	Output *intf.OutputChannel
}

// AdvanceRow will update the current state to make room for the next row's state data
func (cs *ChannelState) AdvanceRow() {
	cs.prevState = cs.activeState
	cs.targetState = cs.activeState.PlaybackState
	cs.DoRetriggerNote = false
	cs.NotePlayTick = 0
	cs.RetriggerCount = 0
	cs.activeState.PeriodDelta = 0

	cs.WantNoteCalc = false
	cs.WantVolCalc = false
	cs.UseTargetPeriod = false
}

// RenderRowTick renders a channel's row data for a single tick
func (cs *ChannelState) RenderRowTick(mix *mixing.Mixer, panmixer mixing.PanMixer, samplerSpeed float32, tickSamples int, tickDuration time.Duration) (*mixing.Data, error) {
	if cs.PlaybackFrozen() {
		return nil, nil
	}

	var (
		mixData *mixing.Data
		err     error
	)
	if cs.activeState.Enabled {
		mixData, err = cs.activeState.Render(mix, panmixer, samplerSpeed, tickSamples, tickDuration)
		if mixData == nil {
			cs.activeState.Enabled = false
		}
	}
	if err != nil {
		return mixData, err
	}
	var uNotes []*activeState
	for _, pn := range cs.pastNote {
		if pn.Enabled {
			if pn.NoteControl != nil && pn.Period != nil {
				ps, err2 := pn.Render(mix, panmixer, samplerSpeed, tickSamples, tickDuration)
				if ps == nil {
					pn.Enabled = false
				}
				if err == nil && err2 != nil {
					err = err2
				}
				if ps != nil && ps.Data != nil {
					if mixData == nil || mixData.Data == nil {
						mixData = ps
					} else {
						centerPan := ps.Volume.Apply(panmixer.GetMixingMatrix(panning.CenterAhead)...)
						mixData.Data.Add(0, ps.Data, centerPan)
					}
				}
			} else {
				pn.Enabled = false
			}
		}
		if pn.Enabled {
			uNotes = append(uNotes, pn)
		}
	}
	cs.pastNote = uNotes
	return mixData, err
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

// SetPeriodDelta sets the vibrato (ephemeral) delta sampler period
func (cs *ChannelState) SetPeriodDelta(delta note.PeriodDelta) {
	cs.activeState.PeriodDelta = delta
}

// GetPeriodDelta gets the vibrato (ephemeral) delta sampler period
func (cs *ChannelState) GetPeriodDelta() note.PeriodDelta {
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
func (cs *ChannelState) SetInstrument(inst intf.Instrument) {
	cs.activeState.Instrument = inst
	if cs.prevState.Instrument != inst {
		if prevNc := cs.prevState.NoteControl; prevNc != nil && prevNc.GetKeyOn() {
			prevNc.Release()
		}
	}
	if inst != nil {
		cs.activeState.Enabled = true
		if inst == cs.prevState.Instrument {
			cs.activeState.NoteControl = cs.prevState.NoteControl
		} else {
			cs.activeState.NoteControl = inst.InstantiateOnChannel(cs.Output)
		}
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

// GetPrevInst returns the interface to the last row's active instrument
func (cs *ChannelState) GetPrevInst() intf.Instrument {
	return cs.prevState.Instrument
}

// GetPrevNoteControl returns the interface to the last row's active note-control
func (cs *ChannelState) GetPrevNoteControl() intf.NoteControl {
	return cs.prevState.NoteControl
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

// SetTargetSemitone sets the target semitone for the channel
func (cs *ChannelState) SetTargetSemitone(st note.Semitone) {
	cs.TargetSemitone = st
}

// SetStoredSemitone sets the stored semitone for the channel
func (cs *ChannelState) SetStoredSemitone(st note.Semitone) {
	cs.StoredSemitone = st
}

// SetOutputChannel sets the output channel for the channel
func (cs *ChannelState) SetOutputChannel(outputCh *intf.OutputChannel) {
	cs.Output = outputCh
}

// GetOutputChannel returns the output channel for the channel
func (cs *ChannelState) GetOutputChannel() *intf.OutputChannel {
	return cs.Output
}

// SetGlobalVolume sets the last-known global volume on the channel
func (cs *ChannelState) SetGlobalVolume(gv volume.Volume) {
	cs.Output.GlobalVolume = gv
}

// SetChannelVolume sets the channel volume on the channel
func (cs *ChannelState) SetChannelVolume(cv volume.Volume) {
	cs.Output.ChannelVolume = cv
}

// GetChannelVolume gets the channel volume on the channel
func (cs *ChannelState) GetChannelVolume() volume.Volume {
	return cs.Output.ChannelVolume
}

// SetEnvelopePosition sets the envelope position for the active instrument
func (cs *ChannelState) SetEnvelopePosition(ticks int) {
	if nc := cs.GetNoteControl(); nc != nil {
		nc.SetEnvelopePosition(ticks)
	}
}

// TransitionActiveToPastState will transition the current active state to the 'past' state
// and will activate the specified New-Note Action on it
func (cs *ChannelState) TransitionActiveToPastState() {
	cs.activeState.NoteControl = nil
	cs.activeState.Instrument = nil
	cs.activeState.Period = nil

	if cs.NewNoteAction == note.ActionNoteCut {
		return
	}

	pn := cs.activeState.Clone()

	switch cs.NewNoteAction {
	//case note.NewNoteActionNoteCut:
	//	pn.Enabled = false
	case note.ActionContinue:
		// nothing
	case note.ActionNoteOff:
		if nc := pn.NoteControl; nc != nil {
			nc.Release()
		}
	case note.ActionFadeout:
		if nc := pn.NoteControl; nc != nil {
			nc.Release()
			nc.Fadeout()
		}
	}
	cs.pastNote = append(cs.pastNote, &pn)
}

// DoPastNoteEffect performs an action on all past-note playbacks associated with the channel
func (cs *ChannelState) DoPastNoteEffect(action note.Action) {
	switch action {
	case note.ActionNoteCut:
		cs.pastNote = nil
	case note.ActionContinue:
		// nothing
	case note.ActionNoteOff:
		for _, pn := range cs.pastNote {
			if nc := pn.NoteControl; nc != nil {
				nc.Release()
			}
		}
	case note.ActionFadeout:
		for _, pn := range cs.pastNote {
			if nc := pn.NoteControl; nc != nil {
				nc.Release()
				nc.Fadeout()
			}
		}
	}
}

// SetNewNoteAction sets the New-Note Action on the channel
func (cs *ChannelState) SetNewNoteAction(nna note.Action) {
	cs.NewNoteAction = nna
}

// GetNewNoteAction gets the New-Note Action on the channel
func (cs *ChannelState) GetNewNoteAction() note.Action {
	return cs.NewNoteAction
}

// SetVolumeEnvelopeEnable sets the enable flag on the active volume envelope
func (cs *ChannelState) SetVolumeEnvelopeEnable(enabled bool) {
	if nc := cs.activeState.NoteControl; nc != nil {
		nc.SetVolumeEnvelopeEnable(enabled)
	}
}

// SetPanningEnvelopeEnable sets the enable flag on the active panning envelope
func (cs *ChannelState) SetPanningEnvelopeEnable(enabled bool) {
	if nc := cs.activeState.NoteControl; nc != nil {
		nc.SetPanningEnvelopeEnable(enabled)
	}
}

// SetPitchEnvelopeEnable sets the enable flag on the active pitch/filter envelope
func (cs *ChannelState) SetPitchEnvelopeEnable(enabled bool) {
	if nc := cs.activeState.NoteControl; nc != nil {
		nc.SetPitchEnvelopeEnable(enabled)
	}
}
