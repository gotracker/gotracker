package util

import (
	"math"
	"syscall"

	s3mfile "github.com/heucuva/goaudiofile/music/tracked/s3m"
	"github.com/heucuva/gomixing/panning"
	"github.com/heucuva/gomixing/volume"

	"gotracker/internal/player/note"
)

const (
	floatDefaultC2Spd = float32(s3mfile.DefaultC2Spd)
	c2Period          = float32(1712)

	// S3MBaseClock is the base clock speed of S3M files
	S3MBaseClock = floatDefaultC2Spd * c2Period
)

var (
	// DefaultVolume is the default volume value for most everything in S3M format
	DefaultVolume = VolumeFromS3M(s3mfile.DefaultVolume)

	// DefaultPanningLeft is the default panning value for left channels
	DefaultPanningLeft = PanningFromS3M(0x03)
	// DefaultPanning is the default panning value for unconfigured channels
	DefaultPanning = PanningFromS3M(0x08)
	// DefaultPanningRight is the default panning value for right channels
	DefaultPanningRight = PanningFromS3M(0x0C)
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
		c2spd = note.C2SPD(s3mfile.DefaultC2Spd)
	}

	period := (note.Period(floatDefaultC2Spd*semitonePeriodTable[key]) / note.Period(uint32(c2spd)<<octave))
	return period.AddInteger(0)
}

// VolumeFromS3M converts an S3M volume to a player volume
func VolumeFromS3M(vol s3mfile.Volume) volume.Volume {
	var v volume.Volume
	switch {
	case vol == s3mfile.EmptyVolume:
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
func VolumeToS3M(v volume.Volume) s3mfile.Volume {
	switch {
	case v == volume.VolumeUseInstVol:
		return s3mfile.EmptyVolume
	default:
		return s3mfile.Volume(v * 64.0)
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
	return syscall.Ntohs(be)
}

// PanningFromS3M returns a radian panning position from an S3M panning value
func PanningFromS3M(pos uint8) panning.Position {
	prad := float64(pos) * math.Pi / 32.0

	return panning.Position{
		Angle:    float32(prad),
		Distance: 1.0,
	}
}

// NoteFromS3MNote converts an S3M file note into a player note
func NoteFromS3MNote(sn s3mfile.Note) note.Note {
	return note.Note(uint8(sn))
}
