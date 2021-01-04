package util

import "gotracker/internal/player/note"

// LinearPeriod is a linear period, based on semitone and finetune values
type LinearPeriod struct {
	Semitone note.Semitone
	Finetune note.Finetune
	C2Spd    note.C2SPD
}

// Add adds the current period to a delta value then returns the resulting period
func (p *LinearPeriod) Add(delta note.Period) note.Period {
	period := LinearPeriod(*p)
	if d, ok := delta.(*LinearPeriod); ok {
		period.Semitone += d.Semitone
		period.Finetune += d.Finetune
		// ignore c2spd from delta
	}
	return &period
}

// Compare returns:
//  -1 if the current period is higher frequency than the `rhs` period
//  0 if the current period is equal in frequency to the `rhs` period
//  1 if the current period is lower frequency than the `rhs` period
func (p *LinearPeriod) Compare(rhs note.Period) int {
	right := LinearPeriod{
		Semitone: 0,
		Finetune: 0,
		C2Spd:    0,
	}
	if r, ok := rhs.(*LinearPeriod); ok {
		right = *r
	}

	// convert to amiga periods
	lp := CalcLinearPeriod(p.Semitone, p.Finetune, p.C2Spd)
	rp := CalcLinearPeriod(right.Semitone, right.Finetune, right.C2Spd)

	return lp.Compare(rp)
}

// Lerp linear-interpolates the current period with the `rhs` period
func (p *LinearPeriod) Lerp(t float64, rhs note.Period) note.Period {
	var right *LinearPeriod
	if r, ok := rhs.(*LinearPeriod); ok {
		right = r
	} else {
		return p
	}

	period := *p
	period.Semitone += note.Semitone(t * (float64(right.Semitone) - float64(period.Semitone)))
	period.Finetune += note.Finetune(t * (float64(right.Finetune) - float64(period.Finetune)))
	return &period
}

// GetSamplerAdd returns the number of samples to advance an instrument by given the period
func (p *LinearPeriod) GetSamplerAdd(samplerSpeed float64) float64 {
	period := float64(*(CalcLinearPeriod(p.Semitone, p.Finetune, DefaultC2Spd).(*AmigaPeriod)))
	if period == 0 {
		return 0
	}
	return samplerSpeed / period
}
