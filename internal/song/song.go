package song

import (
	"gotracker/internal/song/index"
	"gotracker/internal/song/instrument"
	"gotracker/internal/song/note"
)

// Data is an interface to the song data
type Data interface {
	GetOrderList() []index.Pattern
	GetPattern(index.Pattern) Pattern
	IsChannelEnabled(int) bool
	GetOutputChannel(int) int
	NumInstruments() int
	IsValidInstrumentID(instrument.InstrumentID) bool
	GetInstrument(instrument.InstrumentID) (*instrument.Instrument, note.Semitone)
	GetName() string
}
