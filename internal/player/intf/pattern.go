package intf

import "errors"

var (
	// ErrStopSong is a magic error asking to stop the current song
	ErrStopSong = errors.New("stop song")
)

// Pattern is an interface for pattern data
type Pattern interface {
	GetRow(RowIdx) Row
	GetRows() Rows
}

// Patterns is an array of pattern interfaces
type Patterns []Pattern

// Rows is an interface to obtain row data
type Rows interface {
	GetRow(RowIdx) Row
	NumRows() int
}

// OrderIdx is an index into the pattern order list
type OrderIdx uint8

// PatternIdx is an index into the pattern list
type PatternIdx uint8

// RowIdx is an index into the pattern for the row
type RowIdx uint8

const (
	// NextPattern allows the order system the ability to kick to the next pattern
	NextPattern = PatternIdx(254)
	// InvalidPattern specifies an invalid pattern
	InvalidPattern = PatternIdx(255)
)
