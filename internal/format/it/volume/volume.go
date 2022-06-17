package volume

import (
	itfile "github.com/gotracker/goaudiofile/music/tracked/it"
	"github.com/gotracker/gomixing/volume"
)

var (
	// DefaultVolume is the default volume value for most everything in it format
	DefaultVolume = FromItVolume(0x40)

	// DefaultMixingVolume is the default mixing volume
	DefaultMixingVolume = volume.Volume(0x30) / 0x80
)

// FromItVolume converts an it volume to a player volume
func FromItVolume(vol itfile.Volume) volume.Volume {
	return volume.Volume(vol.Value())
}

// FromVolPan converts an it volume-pan to a player volume
func FromVolPan(vp uint8) volume.Volume {
	switch {
	case vp <= 64:
		return volume.Volume(vp) / 64
	default:
		return volume.VolumeUseInstVol
	}
}

// ToItVolume converts a player volume to an it volume
func ToItVolume(v volume.Volume) itfile.Volume {
	switch {
	case v == volume.VolumeUseInstVol:
		return 0
	default:
		return itfile.Volume(v * 64.0)
	}
}
