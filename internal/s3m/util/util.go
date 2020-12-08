package util

import "gotracker/internal/player/volume"

const (
	DefaultC2Spd = uint16(8363)

	floatDefaultC2Spd = float32(DefaultC2Spd)

	S3MBaseClock = floatDefaultC2Spd * 1712.0
)

var semitonePeriodTable = [...]float32{27392, 25856, 24384, 23040, 21696, 20480, 19328, 18240, 17216, 16256, 15360, 14496}

func CalcSemitonePeriod(semi uint8, c2spd uint16) float32 {
	key := int(semi) % len(semitonePeriodTable)
	octave := uint(int(semi) / len(semitonePeriodTable))

	if key >= len(semitonePeriodTable) {
		return 0
	}

	if c2spd == 0 {
		c2spd = DefaultC2Spd
	}

	return (floatDefaultC2Spd * semitonePeriodTable[key]) / float32(uint32(c2spd)<<octave)
}

func VolumeFromS3M(vol uint8) volume.Volume {
	var v volume.Volume
	switch {
	case vol == 255:
		v = volume.VolumeUseInstVol
	case vol >= 63:
		v = volume.Volume(63.0) / 64.0
	case vol < 63:
		v = volume.Volume(vol) / 64.0
	default:
		v = 0.0
	}
	return v
}

func VolumeToS3M(v volume.Volume) uint8 {
	switch {
	case v == volume.VolumeUseInstVol:
		return 255
	default:
		return uint8(v * 64.0)
	}
}
