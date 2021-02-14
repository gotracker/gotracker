package component

import (
	"gotracker/internal/envelope"
)

// FilterEnvelope is a filter frequency cutoff modulation envelope
type FilterEnvelope struct {
	enabled bool
	state   envelope.State
	value   float32
}

// Reset resets the state to defaults based on the envelope provided
func (e *FilterEnvelope) Reset(env *envelope.Envelope) {
	e.state.Reset(env)
}

// SetEnabled sets the enabled flag for the envelope
func (e *FilterEnvelope) SetEnabled(enabled bool) {
	e.enabled = enabled
}

// IsEnabled returns the enabled flag for the envelope
func (e FilterEnvelope) IsEnabled() bool {
	return e.enabled
}

// GetCurrentValue returns the current cached envelope value
func (e FilterEnvelope) GetCurrentValue() float32 {
	return e.value
}

// Advance advances the envelope state 1 tick and calculates the current envelope value
func (e *FilterEnvelope) Advance(keyOn bool, prevKeyOn bool) {
	e.state.Advance(keyOn, prevKeyOn)
	cur, next, t := e.state.GetCurrentValue(keyOn)

	y0 := float32(0)
	if cur != nil {
		cur.Value(&y0)
	}

	y1 := float32(0)
	if next != nil {
		next.Value(&y1)
	}

	e.value = y0 + t*(y1-y0)
}
