package component

import "gotracker/internal/envelope"

// Envelope is an envelope component interface
type Envelope interface {
	//Reset(env *envelope.Envelope)
	SetEnabled(enabled bool)
	IsEnabled() bool
	Advance(keyOn bool, prevKeyOn bool)
}

type baseEnvelope struct {
	enabled   bool
	state     envelope.State
	keyOn     bool
	prevKeyOn bool
}

// Reset resets the state to defaults based on the envelope provided
func (e *baseEnvelope) Reset(env *envelope.Envelope) {
	e.state.Reset(env)
	e.keyOn = false
	e.prevKeyOn = false
}

// SetEnabled sets the enabled flag for the envelope
func (e *baseEnvelope) SetEnabled(enabled bool) {
	e.enabled = enabled
}

// IsEnabled returns the enabled flag for the envelope
func (e *baseEnvelope) IsEnabled() bool {
	return e.enabled
}

// SetEnvelopePosition sets the current position in the envelope
func (e *baseEnvelope) SetEnvelopePosition(pos int) {
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
func (e *baseEnvelope) Advance(keyOn bool, prevKeyOn bool) {
	e.keyOn = keyOn
	e.prevKeyOn = prevKeyOn
	e.state.Advance(e.keyOn, e.prevKeyOn)
}
