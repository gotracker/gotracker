package intf

import (
	"github.com/heucuva/gomixing/volume"

	"gotracker/internal/module/player/note"
)

// EffectFactoryFunc is a function type that gets an effect for specified channel data
type EffectFactoryFunc func(mi Memory, data ChannelData) Effect

// CalcSemitonePeriodFunc is a function type that returns the period for a specified note & c2spd
type CalcSemitonePeriodFunc func(semi note.Semitone, c2spd note.C2SPD) note.Period

// Song is an interface to the song state
type Song interface {
	SetCurrentOrder(uint8)
	SetCurrentRow(uint8)
	SetTempo(int)
	DecreaseTempo(int)
	IncreaseTempo(int)
	GetGlobalVolume() volume.Volume
	SetGlobalVolume(volume.Volume)
	SetTicks(int)
	AddRowTicks(int)
	SetPatternDelay(int)
	SetPatternLoopStart()
	SetPatternLoopEnd(uint8)
	CanPatternLoop() bool
	SetEffectFactory(EffectFactoryFunc)
	SetCalcSemitonePeriod(CalcSemitonePeriodFunc)
	SetPatterns(Patterns)
	SetOrderList([]uint8)
	SetSongData(SongData)
	SetNumChannels(int)
	GetNumChannels() int
	GetChannel(int) Channel
	GetCurrentOrder() uint8
	GetCurrentRow() uint8
}

// SongData is an interface to the song data
type SongData interface {
	GetOrderList() []uint8
	GetPattern(uint8) Pattern
	IsChannelEnabled(int) bool
	NumInstruments() int
	GetInstrument(int) Instrument
	GetName() string
}
