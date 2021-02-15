package voice

import (
	"github.com/gotracker/gomixing/volume"
)

// AmpModulator is the instrument volume (amplitude) control interface
type AmpModulator interface {
	SetVolume(vol volume.Volume)
	GetVolume() volume.Volume
	GetFinalVolume() volume.Volume
}
