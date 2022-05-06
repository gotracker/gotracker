package song

import (
	"errors"

	"github.com/gotracker/gotracker/internal/song/index"
)

var (
	// ErrStopSong is a magic error asking to stop the current song
	ErrStopSong = errors.New("stop song")
)

// Pattern is an interface for pattern data
type Pattern[TChannelData any] interface {
	GetRow(index.Row) Row[TChannelData]
	GetRows() Rows[TChannelData]
}

// Patterns is an array of pattern interfaces
type Patterns[TChannelData any] []Pattern[TChannelData]

// Rows is an interface to obtain row data
type Rows[TChannelData any] interface {
	GetRow(index.Row) Row[TChannelData]
	NumRows() int
}
