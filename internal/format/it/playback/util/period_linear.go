package util

import (
	"gotracker/internal/player/note"
)

// LinearPeriod is a linear period, based on semitone and finetune values
type LinearPeriod struct {
	Finetune note.Finetune
	C2Spd    note.C2SPD
}

// Add adds the current period to a delta value then returns the resulting period
func (p *LinearPeriod) Add(delta note.PeriodDelta) note.Period {
	period := *p
	// 0 means "not playing", so keep it that way
	if period.Finetune > 0 {
		period.Finetune += note.Finetune(delta)
		if period.Finetune < 1 {
			period.Finetune = 1
		}
	}
	return &period
}

// Compare returns:
//  -1 if the current period is higher frequency than the `rhs` period
//  0 if the current period is equal in frequency to the `rhs` period
//  1 if the current period is lower frequency than the `rhs` period
func (p *LinearPeriod) Compare(rhs note.Period) note.SpaceshipResult {
	lf := p.GetFrequency()
	rf := rhs.GetFrequency()

	switch {
	case lf < rf:
		return note.CompareRightHigher
	case lf > rf:
		return note.CompareLeftHigher
	default:
		return note.CompareEqual
	}
}

// Lerp linear-interpolates the current period with the `rhs` period
func (p *LinearPeriod) Lerp(t float64, rhs note.Period) note.Period {
	right := ToLinearPeriod(rhs)

	period := *p

	lnft := float64(period.Finetune)
	rnft := float64(right.Finetune)

	delta := note.PeriodDelta(t * (rnft - lnft))
	period.Add(delta)
	return &period
}

// GetSamplerAdd returns the number of samples to advance an instrument by given the period
func (p *LinearPeriod) GetSamplerAdd(samplerSpeed float64) float64 {
	period := float64(ToAmigaPeriod(p.Finetune, p.C2Spd))
	if period == 0 {
		return 0
	}
	return samplerSpeed / period
}

// GetFrequency returns the frequency defined by the period
func (p *LinearPeriod) GetFrequency() float64 {
	am := ToAmigaPeriod(p.Finetune, p.C2Spd)
	return am.GetFrequency()
}
