package intf

import (
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/note"
)

// ChannelData is an interface for channel data
type ChannelData interface {
	HasNote() bool
	GetNote() note.Note

	HasInstrument() bool
	GetInstrument(note.Semitone) InstrumentID

	HasVolume() bool
	GetVolume() volume.Volume

	HasCommand() bool

	Channel() uint8
}

// Channel is an interface for channel state
type Channel interface {
	ResetRetriggerCount()
	SetMemory(Memory)
	GetMemory() Memory
	GetActiveVolume() volume.Volume
	SetActiveVolume(volume.Volume)
	FreezePlayback()
	UnfreezePlayback()
	GetData() ChannelData
	GetPortaTargetPeriod() note.Period
	SetPortaTargetPeriod(note.Period)
	GetTargetPeriod() note.Period
	SetTargetPeriod(note.Period)
	GetPeriod() note.Period
	SetPeriod(note.Period)
	SetPeriodDelta(note.PeriodDelta)
	GetPeriodDelta() note.PeriodDelta
	SetInstrument(Instrument)
	GetInstrument() Instrument
	GetNoteControl() NoteControl
	GetTargetInst() Instrument
	SetTargetInst(Instrument)
	GetPrevInst() Instrument
	GetPrevNoteControl() NoteControl
	GetNoteSemitone() note.Semitone
	SetStoredSemitone(note.Semitone)
	SetTargetSemitone(note.Semitone)
	GetTargetPos() sampling.Pos
	SetTargetPos(sampling.Pos)
	GetPos() sampling.Pos
	SetPos(sampling.Pos)
	SetNotePlayTick(int)
	GetRetriggerCount() uint8
	SetRetriggerCount(uint8)
	SetPanEnabled(bool)
	GetPan() panning.Position
	SetPan(panning.Position)
	SetDoRetriggerNote(bool)
	SetOutputChannel(*OutputChannel)
	GetOutputChannel() *OutputChannel
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
