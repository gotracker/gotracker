package component

import (
	"gotracker/internal/envelope"
)

// FilterEnvelope is a filter frequency cutoff modulation envelope
type FilterEnvelope struct {
	enabled   bool
	state     envelope.State
	value     float32
	keyOn     bool
	prevKeyOn bool
}

// Reset resets the state to defaults based on the envelope provided
func (e *FilterEnvelope) Reset(env *envelope.Envelope) {
	e.state.Reset(env)
	e.keyOn = false
	e.prevKeyOn = false
	e.update()
}

// SetEnabled sets the enabled flag for the envelope
func (e *FilterEnvelope) SetEnabled(enabled bool) {
	e.enabled = enabled
}

// IsEnabled returns the enabled flag for the envelope
func (e *FilterEnvelope) IsEnabled() bool {
	return e.enabled
}

// GetCurrentValue returns the current cached envelope value
func (e *FilterEnvelope) GetCurrentValue() float32 {
	return e.value
}

// SetEnvelopePosition sets the current position in the envelope
func (e *FilterEnvelope) SetEnvelopePosition(pos int) {
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
func (e *FilterEnvelope) Advance(keyOn bool, prevKeyOn bool) {
	e.keyOn = keyOn
	e.prevKeyOn = prevKeyOn
	e.state.Advance(e.keyOn, e.prevKeyOn)
	e.update()
}

func (e *FilterEnvelope) update() {
	cur, next, t := e.state.GetCurrentValue(e.keyOn)

	y0 := float32(0)
	if cur != nil {
		cur.Value(&y0)
	}

	y1 := float32(0)
	if next != nil {
		next.Value(&y1)
	}

	e.value = y0 + t*(y1-y0)
	e.value /= 256
}
