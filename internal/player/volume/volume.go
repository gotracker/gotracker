package volume

import "math"

// Volume is a mixable volume
type Volume float32

var (
	// VolumeUseInstVol tells the system to use the volume stored on the instrument
	VolumeUseInstVol = Volume(math.Inf(-1))
)

type uint24 struct {
	Hi uint8
	Lo uint16
}

// ToSample returns a volume as a typed value supporting the bits per sample provided
func (v Volume) ToSample(bitsPerSample int) interface{} {
	switch bitsPerSample {
	case 8:
		return uint8(v * 128.0)
	case 16:
		return uint16(v * 32678.0)
	case 24:
		s := uint32(v * 8388608.0)
		return uint24{Hi: uint8(s >> 16), Lo: uint16(s & 65535)}
	case 32:
		return uint32(v * 2147483648.0)
	}
	return 0
}
