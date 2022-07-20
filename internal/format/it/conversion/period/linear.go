package period

import (
	"fmt"
	"math"

	"github.com/gotracker/gotracker/internal/comparison"
	"github.com/gotracker/gotracker/internal/song/note"

	"github.com/gotracker/voice/period"
)

// Linear is a linear period, based on semitone and finetune values
type Linear struct {
	Finetune note.Finetune
	C2Spd    note.C2SPD
}

// Add adds the current period to a delta value then returns the resulting period
func (p Linear) AddDelta(delta period.Delta) period.Period {
	period := p
	// 0 means "not playing", so keep it that way
	if period.Finetune > 0 {
		d := note.ToPeriodDelta(delta)
		period.Finetune += note.Finetune(d)
		if period.Finetune < 1 {
			period.Finetune = 1
		}
	}
	return period
}

// Compare returns:
//  -1 if the current period is higher frequency than the `rhs` period
//  0 if the current period is equal in frequency to the `rhs` period
//  1 if the current period is lower frequency than the `rhs` period
func (p Linear) Compare(rhs note.Period) comparison.Spaceship {
	lf := p.GetFrequency()
	rf := rhs.GetFrequency()

	switch {
	case lf < rf:
		return comparison.SpaceshipRightGreater
	case lf > rf:
		return comparison.SpaceshipLeftGreater
	default:
		return comparison.SpaceshipEqual
	}
}

// Lerp linear-interpolates the current period with the `rhs` period
func (p Linear) Lerp(t float64, rhs note.Period) note.Period {
	right := ToLinearPeriod(rhs)

	period := p

	lnft := float64(period.Finetune)
	rnft := float64(right.Finetune)

	delta := note.PeriodDelta(t * (rnft - lnft))
	period.AddDelta(delta)
	return period
}

// GetSamplerAdd returns the number of samples to advance an instrument by given the period
func (p Linear) GetSamplerAdd(samplerSpeed float64) float64 {
	period := float64(ToAmigaPeriod(p.Finetune, p.C2Spd))
	if period == 0 {
		return 0
	}
	return samplerSpeed / period
}

// GetFrequency returns the frequency defined by the period
func (p Linear) GetFrequency() period.Frequency {
	am := ToAmigaPeriod(p.Finetune, p.C2Spd)
	return am.GetFrequency()
}

func (p Linear) String() string {
	return fmt.Sprintf("Linear{ Finetune:%v C2Spd:%v }", p.Finetune, p.C2Spd)
}

// ToLinearPeriod returns the linear frequency period for a given period
func ToLinearPeriod(p note.Period) Linear {
	switch pp := p.(type) {
	case Linear:
		return pp
	case Amiga:
		linFreq := float64(semitonePeriodTable[0]) / float64(pp)

		fts := note.Finetune(semitonesPerOctave * math.Log2(linFreq))

		lp := Linear{
			Finetune: fts,
			C2Spd:    DefaultC2Spd,
		}
		return lp
	}
	return Linear{}
}
