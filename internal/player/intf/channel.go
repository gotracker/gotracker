package intf

import (
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/note"
	"gotracker/internal/player/oscillator"
)

// ChannelData is an interface for channel data
type ChannelData interface {
	HasNote() bool
	GetNote() note.Note

	HasInstrument() bool
	GetInstrument() InstrumentID

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
	SetStoredVolume(volume.Volume, volume.Volume)
	FreezePlayback()
	UnfreezePlayback()
	GetData() ChannelData
	GetPortaTargetPeriod() note.Period
	SetPortaTargetPeriod(note.Period)
	GetTargetPeriod() note.Period
	SetTargetPeriod(note.Period)
	GetPeriod() note.Period
	SetPeriod(note.Period)
	SetVibratoDelta(note.Period)
	GetVibratoOscillator() *oscillator.Oscillator
	GetTremoloOscillator() *oscillator.Oscillator
	GetTremorOn() bool
	SetTremorOn(bool)
	GetTremorTime() int
	SetTremorTime(int)
	SetInstrument(Instrument)
	GetInstrument() Instrument
	GetTargetInst() Instrument
	SetTargetInst(Instrument)
	GetNoteSemitone() note.Semitone
	SetSemitone(note.Semitone)
	GetTargetPos() sampling.Pos
	SetTargetPos(sampling.Pos)
	GetPos() sampling.Pos
	SetPos(sampling.Pos)
	SetNotePlayTick(int)
	GetRetriggerCount() uint8
	SetRetriggerCount(uint8)
	SetPan(panning.Position)
	SetDoRetriggerNote(bool)
	GetFilter() Filter
	SetFilter(Filter)
	SetOutputChannelNum(int)
}
