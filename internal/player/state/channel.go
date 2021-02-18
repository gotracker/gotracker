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

// NoteTriggerDetails is for when a note needs to be played
type NoteTriggerDetails struct {
	Tick int
}

// ChannelState is the state of a single channel
type ChannelState struct {
	activeState ActiveState
	targetState intf.PlaybackState
	prevState   ActiveState

	ActiveEffect intf.Effect

	TargetSemitone note.Semitone // from pattern, modified

	StoredSemitone    note.Semitone // from pattern, unmodified, current note
	PortaTargetPeriod note.Period
	Trigger           *NoteTriggerDetails
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
	pastNote          []*ActiveState

	Output *intf.OutputChannel
}

// WillTriggerOn returns true if a note will trigger on the tick specified
func (cs *ChannelState) WillTriggerOn(tick int) bool {
	if cs.Trigger == nil {
		return false
	}

	return cs.Trigger.Tick == tick
}

// AdvanceRow will update the current state to make room for the next row's state data
func (cs *ChannelState) AdvanceRow() {
	cs.prevState = cs.activeState
	cs.targetState = cs.activeState.PlaybackState
	cs.Trigger = nil
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

	activeStates := []*ActiveState{&cs.activeState}
	activeStates = append(activeStates, cs.pastNote...)
	mixData, participatingStates := RenderStatesTogether(activeStates, mix, panmixer, samplerSpeed, tickSamples, tickDuration)

	var uNotes []*ActiveState
	for _, pn := range participatingStates {
		if pn != &cs.activeState {
			uNotes = append(uNotes, pn)
		}
	}
	cs.pastNote = uNotes
	return mixData, nil
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
			cs.activeState.NoteControl = cs.newNoteControl()
		}
	}
}

// newNoteControl takes an instrument and loads it onto an output channel
func (cs *ChannelState) newNoteControl() intf.NoteControl {
	ioc := NoteControl{
		Output: cs.Output,
	}

	if inst := cs.activeState.Instrument; inst != nil {
		ioc.SetupVoice(inst)

		if cfact := inst.GetChannelFilterFactory(); cfact != nil {
			ioc.Output.Filter = cfact(ioc.Output.Playback.GetSampleRate())
		}
	}

	return &ioc
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
func (cs *ChannelState) SetNotePlayTick(enabled bool, tick int) {
	if !enabled {
		cs.Trigger = nil
		return
	}

	if cs.Trigger == nil {
		cs.Trigger = &NoteTriggerDetails{}
	}
	cs.Trigger.Tick = tick
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

// SetTargetSemitone sets the target semitone for the channel
func (cs *ChannelState) SetTargetSemitone(st note.Semitone) {
	cs.TargetSemitone = st
	cs.WantNoteCalc = true
	cs.UseTargetPeriod = true
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
	defer func() {
		cs.activeState.NoteControl = nil
		cs.activeState.Instrument = nil
		cs.activeState.Period = nil
	}()

	if cs.NewNoteAction == note.ActionNoteCut {
		return
	}

	// TODO: This code should be active, but right now it's chewing CPU like mad
	/*
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
		if len(cs.pastNote) > 2 {
			cs.pastNote = cs.pastNote[len(cs.pastNote)-2:]
		}
	*/
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
