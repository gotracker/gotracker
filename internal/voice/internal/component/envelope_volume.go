package component

import (
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/envelope"
)

// VolumeEnvelope is an amplitude modulation envelope
type VolumeEnvelope struct {
	enabled bool
	state   envelope.State
	vol     volume.Volume
}

// Reset resets the state to defaults based on the envelope provided
func (e *VolumeEnvelope) Reset(env *envelope.Envelope) {
	e.state.Reset(env)
}

// SetEnabled sets the enabled flag for the envelope
func (e *VolumeEnvelope) SetEnabled(enabled bool) {
	e.enabled = enabled
}

// IsEnabled returns the enabled flag for the envelope
func (e VolumeEnvelope) IsEnabled() bool {
	return e.enabled
}

// GetCurrentValue returns the current cached envelope value
func (e VolumeEnvelope) GetCurrentValue() volume.Volume {
	return e.vol
}

// Advance advances the envelope state 1 tick and calculates the current envelope value
func (e *VolumeEnvelope) Advance(keyOn bool, prevKeyOn bool) {
	e.state.Advance(keyOn, prevKeyOn)
	cur, next, t := e.state.GetCurrentValue(keyOn)

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
