package voice

import (
	"github.com/gotracker/gomixing/volume"
)

// VolumeEnveloper is a volume envelope interface
type VolumeEnveloper interface {
	EnableVolumeEnvelope(enabled bool)
	IsVolumeEnvelopeEnabled() bool
	GetCurrentVolumeEnvelope() volume.Volume
	SetVolumeEnvelopePosition(pos int)
}
