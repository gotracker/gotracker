package intf

import (
	"gotracker/internal/player/note"
	"gotracker/internal/player/oscillator"
	"gotracker/internal/player/volume"
)

type ChannelData interface {
	HasNote() bool
	GetNote() note.Note

	HasInstrument() bool
	GetInstrument() uint8

	HasVolume() bool
	GetVolume() volume.Volume

	HasCommand() bool

	Channel() uint8
}

type Channel interface {
	ResetRetriggerCount()
	GetMemory() Memory
	SetEffectSharedMemoryIfNonZero(uint8)
	GetEffectSharedMemory(uint8) uint8
	GetActiveVolume() volume.Volume
	SetActiveVolume(volume.Volume)
	FreezePlayback()
	UnfreezePlayback()
	GetData() ChannelData
	GetPortaTargetPeriod() float32
	SetPortaTargetPeriod(float32)
	GetTargetPeriod() float32
	SetTargetPeriod(float32)
	GetPeriod() float32
	SetPeriod(float32)
	SetVibratoDelta(float32)
	GetVibratoOscillator() *oscillator.Oscillator
	GetTremoloOscillator() *oscillator.Oscillator
	GetTremorOn() bool
	SetTremorOn(bool)
	GetTremorTime() int
	SetTremorTime(int)
	GetInstrument() Instrument
	GetTargetInst() Instrument
	SetTargetInst(Instrument)
	GetNoteSemitone() uint8
	SetTargetPos(float32)
	GetPos() float32
	SetPos(float32)
	SetNotePlayTick(int)
	GetRetriggerCount() uint8
	SetRetriggerCount(uint8)
	SetPan(uint8)
}
