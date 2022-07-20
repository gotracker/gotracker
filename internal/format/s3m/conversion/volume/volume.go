package volume

import (
	s3mfile "github.com/gotracker/goaudiofile/music/tracked/s3m"
	"github.com/gotracker/gomixing/volume"
)

var (
	// DefaultVolume is the default volume value for most everything in S3M format
	DefaultVolume = VolumeFromS3M(s3mfile.DefaultVolume)
)

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
