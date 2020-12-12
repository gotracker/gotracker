package util

import (
	"gotracker/internal/player/note"
	"gotracker/internal/player/volume"
)

const (
	// DefaultC2Spd is the default C2SPD for S3M files
	DefaultC2Spd = note.C2SPD(8363)

	floatDefaultC2Spd = float32(DefaultC2Spd)
	c2Period          = float32(1712)

	// S3MBaseClock is the base clock speed of S3M files
	S3MBaseClock = floatDefaultC2Spd * c2Period
)

var semitonePeriodTable = [...]float32{27392, 25856, 24384, 23040, 21696, 20480, 19328, 18240, 17216, 16256, 15360, 14496}

// CalcSemitonePeriod calculates the semitone period for S3M notes
func CalcSemitonePeriod(semi note.Semitone, c2spd note.C2SPD) note.Period {
	key := int(semi) % len(semitonePeriodTable)
	octave := uint(int(semi) / len(semitonePeriodTable))

	if key >= len(semitonePeriodTable) {
		return 0
	}

	if c2spd == 0 {
		c2spd = DefaultC2Spd
	}

	period := (note.Period(floatDefaultC2Spd*semitonePeriodTable[key]) / note.Period(uint32(c2spd)<<octave))
	return period.AddInteger(0)
}

// VolumeFromS3M converts an S3M volume to a player volume
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

// VolumeToS3M converts a player volume to an S3M volume
func VolumeToS3M(v volume.Volume) uint8 {
	switch {
	case v == volume.VolumeUseInstVol:
		return 255
	default:
		return uint8(v * 64.0)
	}
}

// VolumeFromS3M8BitSample converts an S3M 8-bit sample volume to a player volume
func VolumeFromS3M8BitSample(vol uint8) volume.Volume {
	return (volume.Volume(vol) - 128.0) / 128.0
}

// VolumeFromS3M16BitSample converts an S3M 16-bit sample volume to a player volume
func VolumeFromS3M16BitSample(vol uint16) volume.Volume {
	return (volume.Volume(vol) - 32768.0) / 32768.0
}

// BE16ToLE16 converts a big-endian uint16 to a little-endian uint16
func BE16ToLE16(be uint16) uint16 {
	return (be >> 8) | ((be & 0xFF) << 8)
}

// CalcLoopedSamplePos creates a circular buffer of a sample once the position passes the loopEnd position
func CalcLoopedSamplePos(pos float32, loopBegin float32, loopEnd float32) float32 {
	for {
		oldPos := pos
		delta := pos - loopEnd
		if delta < 0 {
			break
		}
		pos = loopBegin + delta
		if pos == oldPos {
			break // don't allow infinite loops
		}
	}
	return pos
}
