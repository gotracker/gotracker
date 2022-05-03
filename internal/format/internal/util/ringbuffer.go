package util

import (
	"golang.org/x/exp/constraints"
)

type RingBuffer[T constraints.Ordered] struct {
	buf  []T
	r    int
	w    int
	full bool
}

func NewRingBuffer[T constraints.Ordered](size int) RingBuffer[T] {
	r := RingBuffer[T]{
		buf:  make([]T, size),
		r:    0,
		w:    0,
		full: false,
	}
	return r
}

func (r *RingBuffer[T]) read(idx int, outVals []T) int {
	if r.w == r.r && !r.full {
		return 0
	}

	size := len(r.buf)

	var n int
	if r.w > idx {
		n = r.w - idx
		if n > len(outVals) {
			n = len(outVals)
		}
		copy(outVals, r.buf[idx:idx+n])
		return n
	}

	n = size - idx + r.w
	if n > len(outVals) {
		n = len(outVals)
	}

	if idx+n <= size {
		copy(outVals, r.buf[idx:idx+n])
	} else {
		c1 := size - idx
		copy(outVals, r.buf[idx:size])
		c2 := n - c1
		copy(outVals[c1:], r.buf[0:c2])
	}

	r.full = false

	return n
}

func (r *RingBuffer[T]) Read(outVals []T) int {
	n := r.read(r.r, outVals)
	r.r = (r.r + n) % len(r.buf)
	return n
}

func (r *RingBuffer[T]) ReadFrom(idx int, outVals []T) int {
	return r.read(idx, outVals)
}

func (r *RingBuffer[T]) Write(vals []T) {
	size := len(r.buf)

	var avail int
	if r.w >= r.r {
		avail = size - r.w + r.r
	} else {
		avail = r.r - r.w
	}

	if len(vals) > avail {
		vals = vals[:avail]
	}
	n := len(vals)

	if r.w >= r.r {
		c1 := size - r.w
		if c1 >= n {
			copy(r.buf[r.w:], vals)
			r.w += n
		} else {
			copy(r.buf[r.w:], vals[:c1])
			c2 := n - c1
			copy(r.buf[0:], vals[c1:])
			r.w = c2
		}
	} else {
		copy(r.buf[r.w:], vals)
		r.w += n
	}

	if r.w == size {
		r.w = 0
	}
	if r.w == r.r {
		r.full = true
	}
}

func (r *RingBuffer[T]) Accumulate(val T) {
	p := r.w - 1
	if p < 0 {
		p += len(r.buf)
	}

	r.buf[p] += val
}
