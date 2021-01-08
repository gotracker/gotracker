package intf

// SongData is an interface to the song data
type SongData interface {
	GetOrderList() []PatternIdx
	GetPattern(PatternIdx) Pattern
	IsChannelEnabled(int) bool
	GetOutputChannel(int) int
	NumInstruments() int
	IsValidInstrumentID(InstrumentID) bool
	GetInstrument(InstrumentID) Instrument
	GetName() string
}

// SongPositionState is an interface to the song position system
type SongPositionState interface {
	AdvanceRow()
	BreakOrder()
	SetNextOrder(OrderIdx)
	SetNextRow(RowIdx, ...bool)
	SetPatternDelay(int)
	SetTempo(int)
	SetTicks(int)
	AccTempoDelta(int)
	SetFinePatternDelay(int)

	GetCurrentOrder() OrderIdx
	GetCurrentRow() RowIdx
}
