package component

import (
	"gotracker/internal/envelope"
	"gotracker/internal/player/note"
)

// PitchEnvelope is an frequency modulation envelope
type PitchEnvelope struct {
	enabled bool
	state   envelope.State
	delta   note.PeriodDelta
}

// Reset resets the state to defaults based on the envelope provided
func (e *PitchEnvelope) Reset(env *envelope.Envelope) {
	e.state.Reset(env)
}

// SetEnabled sets the enabled flag for the envelope
func (e *PitchEnvelope) SetEnabled(enabled bool) {
	e.enabled = enabled
}

// IsEnabled returns the enabled flag for the envelope
func (e PitchEnvelope) IsEnabled() bool {
	return e.enabled
}

// GetCurrentValue returns the current cached envelope value
func (e PitchEnvelope) GetCurrentValue() note.PeriodDelta {
	return e.delta
}

// Advance advances the envelope state 1 tick and calculates the current envelope value
func (e *PitchEnvelope) Advance(keyOn bool, prevKeyOn bool) {
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

	val := y0 + t*(y1-y0)
	e.delta = note.PeriodDelta(val)
}
