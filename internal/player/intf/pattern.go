package intf

import (
	"errors"
	"math"
)

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

// Increment will in-situ increment the value of the index and will return true if an overflow occurs
func (r *RowIdx) Increment(maxRows ...int) bool {
	var overflow bool
	mr := math.MaxUint8
	if len(maxRows) > 0 {
		mr = maxRows[0]
		if mr > 0 {
			mr--
		}
	}
	if int(*r) == mr {
		overflow = true
	}

	*r += 1
	return overflow
}

const (
	// NextPattern allows the order system the ability to kick to the next pattern
	NextPattern = PatternIdx(254)
	// InvalidPattern specifies an invalid pattern
	InvalidPattern = PatternIdx(255)
)
