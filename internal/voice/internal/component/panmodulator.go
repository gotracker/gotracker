package component

import (
	"github.com/gotracker/gomixing/panning"
)

// PanModulator is an pan (spatial) modulator
type PanModulator struct {
	pan panning.Position
}

// SetPan sets the current panning
func (p *PanModulator) SetPan(vol panning.Position) {
	p.pan = vol
}

// GetPan returns the current panning
func (p *PanModulator) GetPan() panning.Position {
	return p.pan
}

// GetFinalPan returns the current panning
func (p *PanModulator) GetFinalPan() panning.Position {
	return p.pan
}

// Advance advances the fadeout value by 1 tick
func (p *PanModulator) Advance() {
}
