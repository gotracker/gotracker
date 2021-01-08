package util

import "gotracker/internal/player/note"

// AmigaPeriod defines a sampler period that follows the Amiga-style approach of note
// definition. Useful in calculating resampling.
type AmigaPeriod float32

// AddInteger truncates the current period to an integer and adds the delta integer in
// then returns the resulting period
func (p *AmigaPeriod) AddInteger(delta int) AmigaPeriod {
	period := AmigaPeriod(int(*p) + delta)
	return period
}

// Add adds the current period to a delta value then returns the resulting period
func (p *AmigaPeriod) Add(delta note.Period) note.Period {
	period := AmigaPeriod(*p)
	if d, ok := delta.(*AmigaPeriod); ok {
		period += *d
	}
	return &period
}

// ToAmigaPeriod returns an Amiga-style period
func (p *AmigaPeriod) ToAmigaPeriod() AmigaPeriod {
	return *p
}

// Compare returns:
//  -1 if the current period is higher frequency than the `rhs` period
//  0 if the current period is equal in frequency to the `rhs` period
//  1 if the current period is lower frequency than the `rhs` period
func (p *AmigaPeriod) Compare(rhs note.Period) int {
	right := AmigaPeriod(0)
	if r, ok := rhs.(*AmigaPeriod); ok {
		right = *r
	}

	switch {
	case *p > right:
		return -1
	case *p < right:
		return 1
	default:
		return 0
	}
}

// Lerp linear-interpolates the current period with the `rhs` period
func (p *AmigaPeriod) Lerp(t float64, rhs note.Period) note.Period {
	right := AmigaPeriod(0)
	if r, ok := rhs.(*AmigaPeriod); ok {
		right = *r
	}

	period := *p
	period += AmigaPeriod(t * (float64(right) - float64(period)))
	return &period
}

// GetSamplerAdd returns the number of samples to advance an instrument by given the period
func (p *AmigaPeriod) GetSamplerAdd(samplerSpeed float64) float64 {
	period := float64(*p)
	if period == 0 {
		return 0
	}
	return samplerSpeed / period
}

// GetFrequency returns the frequency defined by the period
func (p *AmigaPeriod) GetFrequency() float64 {
	return float64(XMBaseClock / float32(*p))
}
