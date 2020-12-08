package volume

import "math"

type Volume float32

var (
	VolumeUseInstVol = Volume(math.Inf(-1))
)

func (v Volume) ToSample(bitsPerSample int) uint32 {
	switch bitsPerSample {
	case 8:
		return uint32(v * 128.0)
	case 16:
		return uint32(v * 32678.0)
	case 24:
		return uint32(v * 8388608.0)
	case 32:
		return uint32(v * 2147483648.0)
	}
	return 0
}
