package intf

import (
	"s3mplayer/internal/player/channel"
	"s3mplayer/internal/player/instrument"
	"s3mplayer/internal/player/oscillator"
)

type Channel interface {
	ResetRetriggerCount()
	GetMemory() Memory
	SetEffectSharedMemoryIfNonZero(uint8)
	GetEffectSharedMemory(uint8) uint8
	GetActiveVolume() uint8
	SetActiveVolume(uint8)
	FreezePlayback()
	UnfreezePlayback()
	GetData() *channel.Data
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
	GetInstrument() *instrument.InstrumentInfo
	GetTargetInst() *instrument.InstrumentInfo
	SetTargetInst(*instrument.InstrumentInfo)
	GetNoteSemitone() uint8
	SetTargetPos(float32)
	GetPos() float32
	SetPos(float32)
	SetNotePlayTick(int)
	GetRetriggerCount() uint8
	SetRetriggerCount(uint8)
	SetPan(uint8)
}
