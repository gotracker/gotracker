package intf

import (
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/note"
)

// EffectFactoryFunc is a function type that gets an effect for specified channel data
type EffectFactoryFunc func(mi Memory, data ChannelData) Effect

// CalcSemitonePeriodFunc is a function type that returns the period for a specified note & c2spd
type CalcSemitonePeriodFunc func(semi note.Semitone, c2spd note.C2SPD) note.Period

// Song is an interface to the song state
type Song interface {
	SetCurrentOrder(OrderIdx)
	SetCurrentRow(RowIdx)
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
	SetOrderList([]PatternIdx)
	SetSongData(SongData)
	GetSongData() SongData
	SetNumChannels(int)
	GetNumChannels() int
	GetChannel(int) Channel
	GetCurrentOrder() OrderIdx
	GetCurrentRow() RowIdx
}

// SongData is an interface to the song data
type SongData interface {
	GetOrderList() []PatternIdx
	GetPattern(PatternIdx) Pattern
	IsChannelEnabled(int) bool
	GetOutputChannel(int) int
	NumInstruments() int
	GetInstrument(int) Instrument
	GetName() string
}
