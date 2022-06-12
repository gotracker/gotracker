package index

import (
	"math"
)

// Row is an index into the pattern for the row
type Row uint8

// Increment will in-situ increment the value of the index and will return true if an overflow occurs
func (r *Row) Increment(maxRows ...int) bool {
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
