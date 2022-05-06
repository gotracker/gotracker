package memory

import "golang.org/x/exp/constraints"

type Value[T constraints.Integer] struct {
	value T
}

// Coalesce will return the input value unles it's zero, in which case, it will return the memory value
// Memory value will be updated if the input value is non-zero
func (v *Value[T]) Coalesce(input T) T {
	if input == 0 {
		return v.value
	} else {
		v.value = input
	}
	return input
}

// CoalesceXY returns the Coalesce()'ed input and memory, split into high and low nibbles
func (v *Value[T]) CoalesceXY(input T) (T, T) {
	xy := v.Coalesce(input)
	return xy >> 4, xy & 0x0f
}

func (v *Value[T]) Reset() {
	v.value = 0
}
