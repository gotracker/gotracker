package intf

import (
	"errors"
	"gotracker/internal/index"
)

var (
	// ErrStopSong is a magic error asking to stop the current song
	ErrStopSong = errors.New("stop song")
)

// Pattern is an interface for pattern data
type Pattern interface {
	GetRow(index.Row) Row
	GetRows() Rows
}

// Patterns is an array of pattern interfaces
type Patterns []Pattern

// Rows is an interface to obtain row data
type Rows interface {
	GetRow(index.Row) Row
	NumRows() int
}
