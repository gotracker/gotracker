package intf

import "gotracker/internal/player/volume"

// Song is an interface to the song state
type Song interface {
	SetCurrentOrder(uint8)
	SetCurrentRow(uint8)
	SetTempo(int)
	DecreaseTempo(int)
	IncreaseTempo(int)
	SetGlobalVolume(volume.Volume)
	SetTicks(int)
	AddRowTicks(int)
	SetPatternDelay(int)
	SetPatternLoopStart()
	SetPatternLoopEnd(uint8)
	CanPatternLoop() bool
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
