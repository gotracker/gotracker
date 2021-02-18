package component

import (
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/envelope"
)

// VolumeEnvelope is an amplitude modulation envelope
type VolumeEnvelope struct {
	enabled   bool
	state     envelope.State
	vol       volume.Volume
	keyOn     bool
	prevKeyOn bool
}

// Reset resets the state to defaults based on the envelope provided
func (e *VolumeEnvelope) Reset(env *envelope.Envelope) {
	e.state.Reset(env)
	e.keyOn = false
	e.prevKeyOn = false
	e.update()
}

// SetEnabled sets the enabled flag for the envelope
func (e *VolumeEnvelope) SetEnabled(enabled bool) {
	e.enabled = enabled
}

// IsEnabled returns the enabled flag for the envelope
func (e *VolumeEnvelope) IsEnabled() bool {
	return e.enabled
}

// GetCurrentValue returns the current cached envelope value
func (e *VolumeEnvelope) GetCurrentValue() volume.Volume {
	return e.vol
}

// SetEnvelopePosition sets the current position in the envelope
func (e *VolumeEnvelope) SetEnvelopePosition(pos int) {
	keyOn := e.keyOn
	prevKeyOn := e.prevKeyOn
	env := e.state.Envelope()
	e.state.Reset(env)
	// TODO: this is gross, but currently the most optimal way to find the correct position
	for i := 0; i < pos; i++ {
		e.Advance(keyOn, prevKeyOn)
	}
}

// Advance advances the envelope state 1 tick and calculates the current envelope value
func (e *VolumeEnvelope) Advance(keyOn bool, prevKeyOn bool) {
	e.keyOn = keyOn
	e.prevKeyOn = prevKeyOn
	e.state.Advance(e.keyOn, e.prevKeyOn)
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
