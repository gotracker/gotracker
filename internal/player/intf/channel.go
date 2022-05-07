package intf

import (
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"
	"github.com/gotracker/voice"

	"github.com/gotracker/gotracker/internal/player/output"
	"github.com/gotracker/gotracker/internal/song/instrument"
	"github.com/gotracker/gotracker/internal/song/note"
)

// Channel is an interface for channel state
type Channel[TMemory, TChannelData any] interface {
	ResetRetriggerCount()
	SetMemory(*TMemory)
	GetMemory() *TMemory
	GetActiveVolume() volume.Volume
	SetActiveVolume(volume.Volume)
	FreezePlayback()
	UnfreezePlayback()
	GetData() *TChannelData
	GetPortaTargetPeriod() note.Period
	SetPortaTargetPeriod(note.Period)
	GetTargetPeriod() note.Period
	SetTargetPeriod(note.Period)
	GetPeriod() note.Period
	SetPeriod(note.Period)
	SetPeriodDelta(note.PeriodDelta)
	GetPeriodDelta() note.PeriodDelta
	SetInstrument(*instrument.Instrument)
	GetInstrument() *instrument.Instrument
	GetVoice() voice.Voice
	GetTargetInst() *instrument.Instrument
	SetTargetInst(*instrument.Instrument)
	GetPrevInst() *instrument.Instrument
	GetPrevVoice() voice.Voice
	GetNoteSemitone() note.Semitone
	SetStoredSemitone(note.Semitone)
	SetTargetSemitone(note.Semitone)
	GetTargetPos() sampling.Pos
	SetTargetPos(sampling.Pos)
	GetPos() sampling.Pos
	SetPos(sampling.Pos)
	SetNotePlayTick(bool, int)
	GetRetriggerCount() uint8
	SetRetriggerCount(uint8)
	SetPanEnabled(bool)
	GetPan() panning.Position
	SetPan(panning.Position)
	SetOutputChannel(*output.Channel)
	GetOutputChannel() *output.Channel
	SetVolumeActive(bool)
	SetGlobalVolume(volume.Volume)
	SetChannelVolume(volume.Volume)
	GetChannelVolume() volume.Volume
	SetEnvelopePosition(int)
	TransitionActiveToPastState()
	SetNewNoteAction(note.Action)
	GetNewNoteAction() note.Action
	DoPastNoteEffect(action note.Action)
	SetVolumeEnvelopeEnable(bool)
	SetPanningEnvelopeEnable(bool)
	SetPitchEnvelopeEnable(bool)
}
