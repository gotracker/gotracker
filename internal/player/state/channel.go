package state

import (
	"github.com/gotracker/gomixing/mixing"
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"
	"github.com/gotracker/voice"

	"github.com/gotracker/gotracker/internal/optional"
	"github.com/gotracker/gotracker/internal/player/intf"
	"github.com/gotracker/gotracker/internal/player/output"
	"github.com/gotracker/gotracker/internal/song/instrument"
	"github.com/gotracker/gotracker/internal/song/note"
	voiceImpl "github.com/gotracker/gotracker/internal/voice"
)

// ChannelState is the state of a single channel
type ChannelState[TMemory, TChannelData any] struct {
	activeState Active[TChannelData]
	targetState Playback
	prevState   Active[TChannelData]

	TargetSemitone note.Semitone // from pattern, modified

	StoredSemitone    note.Semitone // from pattern, unmodified, current note
	PortaTargetPeriod optional.Value[note.Period]
	Trigger           optional.Value[int]
	RetriggerCount    uint8
	Memory            *TMemory
	freezePlayback    bool
	Semitone          note.Semitone // from TargetSemitone, modified further, used in period calculations
	WantNoteCalc      bool
	WantVolCalc       bool
	UseTargetPeriod   bool
	volumeActive      bool
	PanEnabled        bool
	NewNoteAction     note.Action

	PastNotes *PastNotesProcessor[TChannelData]
	Output    *output.Channel
}

// WillTriggerOn returns true if a note will trigger on the tick specified
func (cs *ChannelState[TMemory, TChannelData]) WillTriggerOn(tick int) bool {
	if triggerTick, ok := cs.Trigger.Get(); ok {
		return triggerTick == tick
	}

	return false
}

// AdvanceRow will update the current state to make room for the next row's state data
func (cs *ChannelState[TMemory, TChannelData]) AdvanceRow() {
	cs.prevState = cs.activeState
	cs.targetState = cs.activeState.Playback
	cs.Trigger.Reset()
	cs.RetriggerCount = 0
	cs.activeState.PeriodDelta = 0

	cs.WantNoteCalc = false
	cs.WantVolCalc = false
	cs.UseTargetPeriod = false
}

// RenderRowTick renders a channel's row data for a single tick
func (cs *ChannelState[TMemory, TChannelData]) RenderRowTick(details RenderDetails, pastNotes []*Active[TChannelData]) ([]mixing.Data, error) {
	if cs.PlaybackFrozen() {
		return nil, nil
	}

	mixData := RenderStatesTogether(&cs.activeState, pastNotes, details)

	return mixData, nil
}

// ResetStates resets the channel's internal states
func (cs *ChannelState[TMemory, TChannelData]) ResetStates() {
	cs.activeState.Reset()
	cs.targetState.Reset()
	cs.prevState.Reset()
}

func (cs *ChannelState[TMemory, TChannelData]) GetActiveEffect() intf.Effect {
	return cs.activeState.ActiveEffect
}

func (cs *ChannelState[TMemory, TChannelData]) SetActiveEffect(e intf.Effect) {
	cs.activeState.ActiveEffect = e
}

func (cs *ChannelState[TMemory, TChannelData]) ProcessEffects(p intf.Playback, currentTick int, lastTick bool) error {
	return intf.DoEffect[TMemory, TChannelData](cs.activeState.ActiveEffect, cs, p, currentTick, lastTick)
}

// FreezePlayback suspends mixer progression on the channel
func (cs *ChannelState[TMemory, TChannelData]) FreezePlayback() {
	cs.freezePlayback = true
}

// UnfreezePlayback resumes mixer progression on the channel
func (cs *ChannelState[TMemory, TChannelData]) UnfreezePlayback() {
	cs.freezePlayback = false
}

// PlaybackFrozen returns true if the mixer progression for the channel is suspended
func (cs ChannelState[TMemory, TChannelData]) PlaybackFrozen() bool {
	return cs.freezePlayback
}

// ResetRetriggerCount sets the retrigger count to 0
func (cs *ChannelState[TMemory, TChannelData]) ResetRetriggerCount() {
	cs.RetriggerCount = 0
}

// GetMemory returns the interface to the custom effect memory module
func (cs *ChannelState[TMemory, TChannelData]) GetMemory() *TMemory {
	return cs.Memory
}

// SetMemory sets the custom effect memory interface
func (cs *ChannelState[TMemory, TChannelData]) SetMemory(mem *TMemory) {
	cs.Memory = mem
}

// GetActiveVolume returns the current active volume on the channel
func (cs *ChannelState[TMemory, TChannelData]) GetActiveVolume() volume.Volume {
	return cs.activeState.Volume
}

// SetActiveVolume sets the active volume on the channel
func (cs *ChannelState[TMemory, TChannelData]) SetActiveVolume(vol volume.Volume) {
	if vol != volume.VolumeUseInstVol {
		cs.activeState.Volume = vol
	}
}

// GetData returns the interface to the current channel song pattern data
func (cs *ChannelState[TMemory, TChannelData]) GetData() *TChannelData {
	return cs.activeState.TrackData
}

func (cs *ChannelState[TMemory, TChannelData]) SetData(cdata *TChannelData) {
	cs.activeState.PrevTrackData = cs.activeState.TrackData
	cs.activeState.TrackData = cdata
}

// GetPortaTargetPeriod returns the current target portamento (to note) sampler period
func (cs *ChannelState[TMemory, TChannelData]) GetPortaTargetPeriod() note.Period {
	if p, ok := cs.PortaTargetPeriod.Get(); ok {
		return p
	}
	return nil
}

// SetPortaTargetPeriod sets the current target portamento (to note) sampler period
func (cs *ChannelState[TMemory, TChannelData]) SetPortaTargetPeriod(period note.Period) {
	if period != nil {
		cs.PortaTargetPeriod.Set(period)
	} else {
		cs.PortaTargetPeriod.Reset()
	}
}

// GetTargetPeriod returns the soon-to-be-committed sampler period (when the note retriggers)
func (cs *ChannelState[TMemory, TChannelData]) GetTargetPeriod() note.Period {
	return cs.targetState.Period
}

// SetTargetPeriod sets the soon-to-be-committed sampler period (when the note retriggers)
func (cs *ChannelState[TMemory, TChannelData]) SetTargetPeriod(period note.Period) {
	cs.targetState.Period = period
}

// SetPeriodDelta sets the vibrato (ephemeral) delta sampler period
func (cs *ChannelState[TMemory, TChannelData]) SetPeriodDelta(delta note.PeriodDelta) {
	cs.activeState.PeriodDelta = delta
}

// GetPeriodDelta gets the vibrato (ephemeral) delta sampler period
func (cs *ChannelState[TMemory, TChannelData]) GetPeriodDelta() note.PeriodDelta {
	return cs.activeState.PeriodDelta
}

// SetVolumeActive enables or disables the sample of the instrument
func (cs *ChannelState[TMemory, TChannelData]) SetVolumeActive(on bool) {
	cs.volumeActive = on
}

// GetInstrument returns the interface to the active instrument
func (cs *ChannelState[TMemory, TChannelData]) GetInstrument() *instrument.Instrument {
	return cs.activeState.Instrument
}

// SetInstrument sets the interface to the active instrument
func (cs *ChannelState[TMemory, TChannelData]) SetInstrument(inst *instrument.Instrument) {
	cs.activeState.Instrument = inst
	if inst != nil {
		if inst == cs.prevState.Instrument {
			cs.activeState.Voice = cs.prevState.Voice
		} else {
			cs.activeState.Voice = voiceImpl.New(inst, cs.Output)
		}
	}
}

// GetVoice returns the active voice interface
func (cs *ChannelState[TMemory, TChannelData]) GetVoice() voice.Voice {
	return cs.activeState.Voice
}

// GetTargetInst returns the interface to the soon-to-be-committed active instrument (when the note retriggers)
func (cs *ChannelState[TMemory, TChannelData]) GetTargetInst() *instrument.Instrument {
	return cs.targetState.Instrument
}

// SetTargetInst sets the soon-to-be-committed active instrument (when the note retriggers)
func (cs *ChannelState[TMemory, TChannelData]) SetTargetInst(inst *instrument.Instrument) {
	cs.targetState.Instrument = inst
}

// GetPrevInst returns the interface to the last row's active instrument
func (cs *ChannelState[TMemory, TChannelData]) GetPrevInst() *instrument.Instrument {
	return cs.prevState.Instrument
}

// GetPrevVoice returns the interface to the last row's active voice
func (cs *ChannelState[TMemory, TChannelData]) GetPrevVoice() voice.Voice {
	return cs.prevState.Voice
}

// GetNoteSemitone returns the note semitone for the channel
func (cs *ChannelState[TMemory, TChannelData]) GetNoteSemitone() note.Semitone {
	return cs.StoredSemitone
}

// GetTargetPos returns the soon-to-be-committed sample position of the instrument
func (cs *ChannelState[TMemory, TChannelData]) GetTargetPos() sampling.Pos {
	return cs.targetState.Pos
}

// SetTargetPos sets the soon-to-be-committed sample position of the instrument
func (cs *ChannelState[TMemory, TChannelData]) SetTargetPos(pos sampling.Pos) {
	cs.targetState.Pos = pos
}

// GetPeriod returns the current sampler period of the active instrument
func (cs *ChannelState[TMemory, TChannelData]) GetPeriod() note.Period {
	return cs.activeState.Period
}

// SetPeriod sets the current sampler period of the active instrument
func (cs *ChannelState[TMemory, TChannelData]) SetPeriod(period note.Period) {
	cs.activeState.Period = period
}

// GetPos returns the sample position of the active instrument
func (cs *ChannelState[TMemory, TChannelData]) GetPos() sampling.Pos {
	return cs.activeState.Pos
}

// SetPos sets the sample position of the active instrument
func (cs *ChannelState[TMemory, TChannelData]) SetPos(pos sampling.Pos) {
	cs.activeState.Pos = pos
}

// SetNotePlayTick sets the tick on which the note will retrigger
func (cs *ChannelState[TMemory, TChannelData]) SetNotePlayTick(enabled bool, tick int) {
	if enabled {
		cs.Trigger.Set(tick)
	} else {
		cs.Trigger.Reset()
	}
}

// GetRetriggerCount returns the current count of the retrigger counter
func (cs *ChannelState[TMemory, TChannelData]) GetRetriggerCount() uint8 {
	return cs.RetriggerCount
}

// SetRetriggerCount sets the current count of the retrigger counter
func (cs *ChannelState[TMemory, TChannelData]) SetRetriggerCount(cnt uint8) {
	cs.RetriggerCount = cnt
}

// SetPanEnabled activates or deactivates the panning. If enabled, then pan updates work (see SetPan)
func (cs *ChannelState[TMemory, TChannelData]) SetPanEnabled(on bool) {
	cs.PanEnabled = on
}

// SetPan sets the active panning value of the channel
func (cs *ChannelState[TMemory, TChannelData]) SetPan(pan panning.Position) {
	if cs.PanEnabled {
		cs.activeState.Pan = pan
	}
}

// GetPan gets the active panning value of the channel
func (cs *ChannelState[TMemory, TChannelData]) GetPan() panning.Position {
	return cs.activeState.Pan
}

// SetTargetSemitone sets the target semitone for the channel
func (cs *ChannelState[TMemory, TChannelData]) SetTargetSemitone(st note.Semitone) {
	cs.TargetSemitone = st
	cs.WantNoteCalc = true
	cs.UseTargetPeriod = true
}

// SetStoredSemitone sets the stored semitone for the channel
func (cs *ChannelState[TMemory, TChannelData]) SetStoredSemitone(st note.Semitone) {
	cs.StoredSemitone = st
}

// SetOutputChannel sets the output channel for the channel
func (cs *ChannelState[TMemory, TChannelData]) SetOutputChannel(outputCh *output.Channel) {
	cs.Output = outputCh
}

// GetOutputChannel returns the output channel for the channel
func (cs *ChannelState[TMemory, TChannelData]) GetOutputChannel() *output.Channel {
	return cs.Output
}

// SetGlobalVolume sets the last-known global volume on the channel
func (cs *ChannelState[TMemory, TChannelData]) SetGlobalVolume(gv volume.Volume) {
	cs.Output.LastGlobalVolume = gv
	cs.Output.Config.SetGlobalVolume(gv)
}

// SetChannelVolume sets the channel volume on the channel
func (cs *ChannelState[TMemory, TChannelData]) SetChannelVolume(cv volume.Volume) {
	cs.Output.ChannelVolume = cv
}

// GetChannelVolume gets the channel volume on the channel
func (cs *ChannelState[TMemory, TChannelData]) GetChannelVolume() volume.Volume {
	return cs.Output.ChannelVolume
}

// SetEnvelopePosition sets the envelope position for the active instrument
func (cs *ChannelState[TMemory, TChannelData]) SetEnvelopePosition(ticks int) {
	if nc := cs.GetVoice(); nc != nil {
		voice.SetVolumeEnvelopePosition(nc, ticks)
		voice.SetPanEnvelopePosition(nc, ticks)
		voice.SetPitchEnvelopePosition(nc, ticks)
		voice.SetFilterEnvelopePosition(nc, ticks)
	}
}

// TransitionActiveToPastState will transition the current active state to the 'past' state
// and will activate the specified New-Note Action on it
func (cs *ChannelState[TMemory, TChannelData]) TransitionActiveToPastState() {
	if cs.PastNotes != nil {
		switch cs.NewNoteAction {
		case note.ActionCut:
			// reset at end

		case note.ActionContinue:
			// nothing
			pn := cs.activeState.Clone()
			if nc := pn.Voice; nc != nil {
				cs.PastNotes.Add(cs.Output.ChannelNum, pn)
			}

		case note.ActionRelease:
			pn := cs.activeState.Clone()
			if nc := pn.Voice; nc != nil {
				nc.Release()
				cs.PastNotes.Add(cs.Output.ChannelNum, pn)
			}

		case note.ActionFadeout:
			pn := cs.activeState.Clone()
			if nc := pn.Voice; nc != nil {
				nc.Release()
				nc.Fadeout()
				cs.PastNotes.Add(cs.Output.ChannelNum, pn)
			}
		}
	}
	cs.activeState.Reset()
}

// DoPastNoteEffect performs an action on all past-note playbacks associated with the channel
func (cs *ChannelState[TMemory, TChannelData]) DoPastNoteEffect(action note.Action) {
	cs.PastNotes.Do(cs.Output.ChannelNum, action)
}

// SetNewNoteAction sets the New-Note Action on the channel
func (cs *ChannelState[TMemory, TChannelData]) SetNewNoteAction(nna note.Action) {
	cs.NewNoteAction = nna
}

// GetNewNoteAction gets the New-Note Action on the channel
func (cs *ChannelState[TMemory, TChannelData]) GetNewNoteAction() note.Action {
	return cs.NewNoteAction
}

// SetVolumeEnvelopeEnable sets the enable flag on the active volume envelope
func (cs *ChannelState[TMemory, TChannelData]) SetVolumeEnvelopeEnable(enabled bool) {
	voice.EnableVolumeEnvelope(cs.activeState.Voice, enabled)
}

// SetPanningEnvelopeEnable sets the enable flag on the active panning envelope
func (cs *ChannelState[TMemory, TChannelData]) SetPanningEnvelopeEnable(enabled bool) {
	voice.EnablePanEnvelope(cs.activeState.Voice, enabled)
}

// SetPitchEnvelopeEnable sets the enable flag on the active pitch/filter envelope
func (cs *ChannelState[TMemory, TChannelData]) SetPitchEnvelopeEnable(enabled bool) {
	voice.EnablePitchEnvelope(cs.activeState.Voice, enabled)
}
