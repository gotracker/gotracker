package intf

import (
	"gotracker/internal/index"
	"gotracker/internal/player/note"
)

// SongData is an interface to the song data
type SongData interface {
	GetOrderList() []index.Pattern
	GetPattern(index.Pattern) Pattern
	IsChannelEnabled(int) bool
	GetOutputChannel(int) int
	NumInstruments() int
	IsValidInstrumentID(InstrumentID) bool
	GetInstrument(InstrumentID) (Instrument, note.Semitone)
	GetName() string
}
