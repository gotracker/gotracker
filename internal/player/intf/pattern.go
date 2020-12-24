package intf

// Pattern is an interface for pattern data
type Pattern interface {
	GetRow(RowIdx) Row
	GetRows() []Row
}

// Patterns is an array of pattern interfaces
type Patterns []Pattern

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
