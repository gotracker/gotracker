package util

import (
	"gotracker/internal/player/intf"
)

type loopDetectSequence struct {
	min intf.RowIdx
	max intf.RowIdx
}

type loopDetectNode map[intf.RowIdx]bool

// LoopDetect is a poorly-optimized, but simple loop detection system for tracked music
type LoopDetect struct {
	orders map[intf.OrderIdx]*loopDetectNode
}

// Observe determines if a particular order+row combination has been observed before and returns true if it has
// it will also add the combination to the detection tree if it has not been observed before.
func (ld *LoopDetect) Observe(ord intf.OrderIdx, row intf.RowIdx) bool {
	n := ld.findOrAddOrder(ord)

	if *n == nil {
		*n = make(loopDetectNode)
	}

	if _, found := (*n)[row]; found {
		return true
	}

	(*n)[row] = true
	return false
}

func (ld *LoopDetect) findOrAddOrder(ord intf.OrderIdx) *loopDetectNode {
	if ld.orders == nil {
		ld.orders = make(map[intf.OrderIdx]*loopDetectNode)
	}

	if n, ok := ld.orders[ord]; ok && n != nil {
		return n
	}

	n := loopDetectNode{}
	ld.orders[ord] = &n

	return &n
}
