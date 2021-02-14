package component

import (
	"gotracker/internal/oscillator"
	"gotracker/internal/player/note"
)

// FreqModulator is a frequency (pitch) modulator
type FreqModulator struct {
	period             note.Period
	delta              note.PeriodDelta
	autoVibratoEnabled bool
	autoVibrato        oscillator.Oscillator
	autoVibratoDepth   float32
	autoVibratoRate    int
	autoVibratoSweep   int // maximum age when oscillator is at max depth (in ticks)
	autoVibratoAge     int // current age of oscillator (in ticks)
}

// SetPeriod sets the current period (before AutoVibrato and Delta calculation)
func (a *FreqModulator) SetPeriod(period note.Period) {
	a.period = period
}

// GetPeriod returns the current period (before AutoVibrato and Delta calculation)
func (a FreqModulator) GetPeriod() note.Period {
	return a.period
}

// SetDelta sets the current period delta (before AutoVibrato calculation)
func (a *FreqModulator) SetDelta(delta note.PeriodDelta) {
	a.delta = delta
}

// GetDelta returns the current period delta (before AutoVibrato calculation)
func (a FreqModulator) GetDelta() note.PeriodDelta {
	return a.delta
}

// SetAutoVibratoEnabled sets the status of the AutoVibrato enablement flag
func (a *FreqModulator) SetAutoVibratoEnabled(enabled bool) {
	a.autoVibratoEnabled = enabled
}

// ConfigureAutoVibrato sets the AutoVibrato oscillator settings
func (a *FreqModulator) ConfigureAutoVibrato(osc oscillator.Oscillator, rate int, depth float32) {
	a.autoVibrato = osc
	a.autoVibratoRate = rate
	a.autoVibratoDepth = depth
}

// ResetAutoVibrato resets the current AutoVibrato
func (a *FreqModulator) ResetAutoVibrato(sweep ...int) {
	if a.autoVibrato != nil {
		a.autoVibrato.Reset(true)
	}

	a.autoVibratoAge = 0

	if sweep != nil {
		a.autoVibratoSweep = sweep[0]
	}
}

// IsAutoVibratoEnabled returns the status of the AutoVibrato enablement flag
func (a FreqModulator) IsAutoVibratoEnabled() bool {
	return a.autoVibratoEnabled
}

// GetFinalPeriod returns the current period (after AutoVibrato and Delta calculation)
func (a FreqModulator) GetFinalPeriod() note.Period {
	p := a.period.Add(a.delta)
	if a.autoVibratoEnabled {
		depth := a.autoVibratoDepth
		if a.autoVibratoSweep > a.autoVibratoAge {
			depth *= float32(a.autoVibratoAge) / float32(a.autoVibratoSweep)
		}
		avDelta := a.autoVibrato.GetWave(depth)
		p = p.Add(note.PeriodDelta(avDelta))
	}

	return p
}

// Advance advances the autoVibrato value by 1 tick
func (a *FreqModulator) Advance() {
	if !a.autoVibratoEnabled {
		return
	}

	a.autoVibrato.Advance(a.autoVibratoRate)
	a.autoVibratoAge++
}
