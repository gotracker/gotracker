package intf

import (
	"gotracker/internal/player/note"
)

// CalcSemitonePeriodFunc is a function type that returns the period for a specified note & c2spd
type CalcSemitonePeriodFunc func(semi note.Semitone, c2spd note.C2SPD) note.Period

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

// SongPositionState is an interface to the song position system
type SongPositionState interface {
	AdvanceRow()
	BreakOrder()
	SetNextOrder(OrderIdx)
	SetNextRow(RowIdx)
	SetPatternLoopStart()
	SetPatternLoopEnd()
	SetPatternLoopCount(int)
	SetPatternDelay(int)
	SetTempo(int)
	SetTicks(int)
	AccTempoDelta(int)
	SetFinePatternDelay(int)

	GetCurrentOrder() OrderIdx
	GetCurrentRow() RowIdx
}
