package volume

import (
	"github.com/gotracker/gomixing/volume"
)

var (
	// DefaultVolume is the default volume value for most everything in xm format
	DefaultVolume = ToVolume(0x10 + 0x40)

	// DefaultMixingVolume is the default mixing volume
	DefaultMixingVolume = volume.Volume(0x30) / 0x80
)

// XmVolume is a helpful converter from the XM range of 0-64 into a volume
type XmVolume uint8

const cVolumeXMCoeff = volume.Volume(1) / 0x40

// Volume returns the volume from the internal format
func (v XmVolume) Volume() volume.Volume {
	return volume.Volume(v) * cVolumeXMCoeff
}

// ToVolumeXM returns the VolumeXM representation of a volume
func ToVolumeXM(v volume.Volume) XmVolume {
	return XmVolume(v * 0x40)
}
