package song

import (
	"github.com/gotracker/gotracker/internal/song/index"
	"github.com/gotracker/gotracker/internal/song/instrument"
	"github.com/gotracker/gotracker/internal/song/note"
)

// Data is an interface to the song data
type Data interface {
	GetOrderList() []index.Pattern
	IsChannelEnabled(int) bool
	GetOutputChannel(int) int
	NumInstruments() int
	IsValidInstrumentID(instrument.ID) bool
	GetInstrument(instrument.ID) (*instrument.Instrument, note.Semitone)
	GetName() string
}

type PatternData[TChannelData any] interface {
	GetPattern(index.Pattern) Pattern[TChannelData]
}
