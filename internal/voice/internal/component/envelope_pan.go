package component

import (
	"github.com/gotracker/gomixing/panning"

	"gotracker/internal/envelope"
)

// PanEnvelope is a spatial modulation envelope
type PanEnvelope struct {
	baseEnvelope
	pan panning.Position
}

// Reset resets the state to defaults based on the envelope provided
func (e *PanEnvelope) Reset(env *envelope.Envelope) {
	e.baseEnvelope.Reset(env)
	e.update()
}

// GetCurrentValue returns the current cached envelope value
func (e *PanEnvelope) GetCurrentValue() panning.Position {
	return e.pan
}

// Advance advances the envelope state 1 tick and calculates the current envelope value
func (e *PanEnvelope) Advance(keyOn bool, prevKeyOn bool) {
	e.baseEnvelope.Advance(e.keyOn, e.prevKeyOn)
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
