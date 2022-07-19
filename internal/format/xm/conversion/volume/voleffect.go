package volume

import "github.com/gotracker/gomixing/volume"

// VolEffect holds the data related to volume and effects from the volume data channel
type VolEffect uint8

// IsVolume returns true if the VolEffect describes a volume value
func (v VolEffect) IsVolume() bool {
	return v == 0x00 || v >= 0x10 && v <= 0x50
}

// Volume returns the value from the volume portion of the range
func (v VolEffect) Volume() volume.Volume {
	if v == 0x00 {
		return volume.VolumeUseInstVol
	}
	return XmVolume(v - 0x10).Volume()
}

// ToVolume converts an xm volume to a player volume
func ToVolume(vol VolEffect) volume.Volume {
	if vol.IsVolume() {
		return vol.Volume()
	}
	panic("unexpected conversion of non-volume value")
}

// ToVolEffect converts a player volume to an xm volume
func ToVolEffect(v volume.Volume) VolEffect {
	switch {
	case v == volume.VolumeUseInstVol:
		return 0
	case v >= 0 && v <= 1:
		return VolEffect(v*0x40) + 0x10
	default:
		panic("volume out of range for conversion")
	}
}
