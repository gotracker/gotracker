package util

import "gotracker/internal/song/index"

type loopDetectNode map[index.Row]struct{}

// LoopDetect is a poorly-optimized, but simple loop detection system for tracked music
type LoopDetect struct {
	orders map[index.Order]*loopDetectNode
}

// Observe determines if a particular order+row combination has been observed before and returns true if it has
// it will also add the combination to the detection tree if it has not been observed before.
func (ld *LoopDetect) Observe(ord index.Order, row index.Row) bool {
	n := ld.findOrAddOrder(ord)

	if *n == nil {
		*n = make(loopDetectNode)
	}

	if _, found := (*n)[row]; found {
		return true
	}

	(*n)[row] = struct{}{}
	return false
}

func (ld *LoopDetect) Reset() {
	ld.orders = nil
}

func (ld *LoopDetect) findOrAddOrder(ord index.Order) *loopDetectNode {
	if ld.orders == nil {
		ld.orders = make(map[index.Order]*loopDetectNode)
	}

	if n, ok := ld.orders[ord]; ok && n != nil {
		return n
	}

	n := loopDetectNode{}
	ld.orders[ord] = &n

	return &n
}
