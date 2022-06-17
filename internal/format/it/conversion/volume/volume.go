package volume

import (
	itfile "github.com/gotracker/goaudiofile/music/tracked/it"
	"github.com/gotracker/gomixing/volume"
)

var (
	MaxItVolume     = itfile.Volume(0x40)
	DefaultItVolume = itfile.DefaultVolume

	// DefaultVolume is the default volume value for most everything in IT format
	DefaultVolume = FromItVolume(DefaultItVolume)

	// DefaultMixingVolume is the default mixing volume
	DefaultMixingVolume = itfile.FineVolume(0x30).Value()
)

// FromItVolume converts an it volume to a player volume
func FromItVolume(vol itfile.Volume) volume.Volume {
	return volume.Volume(vol.Value())
}

// FromVolPan converts an it volume-pan to a player volume
func FromVolPan(vp uint8) volume.Volume {
	switch {
	case vp <= uint8(MaxItVolume):
		return volume.Volume(vp) / volume.Volume(MaxItVolume)
	default:
		return volume.VolumeUseInstVol
	}
}

// ToItVolume converts a player volume to an it volume
func ToItVolume(v volume.Volume) itfile.Volume {
	switch {
	case v == volume.VolumeUseInstVol:
		return 0
	case v < 0.0:
		return 0
	case v > 1.0:
		return MaxItVolume
	default:
		return itfile.Volume(v * volume.Volume(MaxItVolume))
	}
}
