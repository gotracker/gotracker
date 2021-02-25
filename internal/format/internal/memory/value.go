package memory

type UInt8 uint8

// Coalesce will return the input value unles it's zero, in which case, it will return the memory value
// Memory value will be updated if the input value is non-zero
func (v *UInt8) Coalesce(input uint8) uint8 {
	if input == 0 {
		return uint8(*v)
	}
	if input != 0 {
		*v = UInt8(input)
	}
	return input
}

// CoalesceXY returns the Coalesce()'ed input and memory, split into high and low nibbles
func (v *UInt8) CoalesceXY(input uint8) (uint8, uint8) {
	xy := v.Coalesce(input)
	return xy >> 4, xy & 0x0f
}
