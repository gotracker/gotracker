package component

import (
	"github.com/gotracker/gomixing/panning"

	"gotracker/internal/envelope"
)

// PanEnvelope is a spatial modulation envelope
type PanEnvelope struct {
	enabled   bool
	state     envelope.State
	pan       panning.Position
	keyOn     bool
	prevKeyOn bool
}

// Reset resets the state to defaults based on the envelope provided
func (e *PanEnvelope) Reset(env *envelope.Envelope) {
	e.state.Reset(env)
	e.keyOn = false
	e.prevKeyOn = false
}

// SetEnabled sets the enabled flag for the envelope
func (e *PanEnvelope) SetEnabled(enabled bool) {
	e.enabled = enabled
}

// IsEnabled returns the enabled flag for the envelope
func (e *PanEnvelope) IsEnabled() bool {
	return e.enabled
}

// GetCurrentValue returns the current cached envelope value
func (e *PanEnvelope) GetCurrentValue() panning.Position {
	return e.pan
}

// SetEnvelopePosition sets the current position in the envelope
func (e *PanEnvelope) SetEnvelopePosition(pos int) {
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
func (e *PanEnvelope) Advance(keyOn bool, prevKeyOn bool) {
	e.keyOn = keyOn
	e.prevKeyOn = prevKeyOn
	e.state.Advance(e.keyOn, e.prevKeyOn)
	e.update()
}

func (e *PanEnvelope) update() {
	cur, next, t := e.state.GetCurrentValue(e.keyOn)

	y0 := panning.CenterAhead
	if cur != nil {
		cur.Value(&y0)
	}

	y1 := panning.CenterAhead
	if next != nil {
		next.Value(&y1)
	}

	// TODO: perform an angular interpolation instead of a linear one.
	e.pan.Angle = y0.Angle + t*(y1.Angle-y0.Angle)
	e.pan.Distance = y0.Distance + t*(y1.Distance-y0.Distance)
}
