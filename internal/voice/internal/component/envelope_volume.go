package component

import (
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/envelope"
)

// VolumeEnvelope is an amplitude modulation envelope
type VolumeEnvelope struct {
	baseEnvelope
	vol volume.Volume
}

// Reset resets the state to defaults based on the envelope provided
func (e *VolumeEnvelope) Reset(env *envelope.Envelope) {
	e.baseEnvelope.Reset(env)
	e.update()
}

// GetCurrentValue returns the current cached envelope value
func (e *VolumeEnvelope) GetCurrentValue() volume.Volume {
	return e.vol
}

// Advance advances the envelope state 1 tick and calculates the current envelope value
func (e *VolumeEnvelope) Advance(keyOn bool, prevKeyOn bool) {
	e.baseEnvelope.Advance(e.keyOn, e.prevKeyOn)
	e.update()
}

func (e *VolumeEnvelope) update() {
	cur, next, t := e.state.GetCurrentValue(e.keyOn)

	y0 := volume.Volume(0)
	if cur != nil {
		cur.Value(&y0)
	}

	y1 := volume.Volume(0)
	if next != nil {
		next.Value(&y1)
	}

	e.vol = y0 + volume.Volume(t)*(y1-y0)
}
