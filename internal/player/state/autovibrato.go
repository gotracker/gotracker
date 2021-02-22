package state

import (
	"github.com/gotracker/voice/oscillator"
)

// AutoVibrato is the information needed to make an instrument auto-vibrato
type AutoVibrato struct {
	Osc   oscillator.Oscillator
	Ticks int
}

// Reset sets the auto-vibrato state to defaults
func (av *AutoVibrato) Reset() {
	if av.Osc != nil {
		av.Osc.Reset()
	}
	av.Ticks = 0
}
