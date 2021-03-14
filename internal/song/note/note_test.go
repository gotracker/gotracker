package note_test

import (
	"fmt"
	"testing"

	"gotracker/internal/comparison"
	"gotracker/internal/song/note"

	"github.com/gotracker/voice/period"
)

// testPeriod defines a sampler period that follows the Amiga-style approach of note
// definition. Useful in calculating resampling.
type testPeriod float32

// AddInteger truncates the current period to an integer and adds the delta integer in
// then returns the resulting period
func (p *testPeriod) AddInteger(delta int) testPeriod {
	period := testPeriod(int(*p) + delta)
	return period
}

// Add adds the current period to a delta value then returns the resulting period
func (p *testPeriod) AddDelta(delta period.Delta) period.Period {
	period := *p
	d := note.ToPeriodDelta(delta)
	period += testPeriod(d)
	return &period
}

// Compare returns:
//  -1 if the current period is higher frequency than the `rhs` period
//  0 if the current period is equal in frequency to the `rhs` period
//  1 if the current period is lower frequency than the `rhs` period
func (p *testPeriod) Compare(rhs note.Period) comparison.Spaceship {
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
func (p *testPeriod) Lerp(t float64, rhs note.Period) note.Period {
	right := testPeriod(0)
	if r, ok := rhs.(*testPeriod); ok {
		right = *r
	}

	period := *p
	delta := note.PeriodDelta(t * (float64(right) - float64(period)))
	period.AddDelta(delta)
	return &period
}

// GetSamplerAdd returns the number of samples to advance an instrument by given the period
func (p *testPeriod) GetSamplerAdd(samplerSpeed float64) float64 {
	period := float64(*p)
	if period == 0 {
		return 0
	}
	return samplerSpeed / period
}

// GetFrequency returns the frequency defined by the period
func (p *testPeriod) GetFrequency() period.Frequency {
	return period.Frequency(p.GetSamplerAdd(float64(8363 * 1712)))
}

func (p *testPeriod) String() string {
	return fmt.Sprintf("%f", *p)
}

func periodCompareTest(t *testing.T, lhs note.Period, rhs note.Period, expected comparison.Spaceship) {
	t.Helper()

	if note.ComparePeriods(lhs, rhs) != expected {
		t.Fatalf("%v <=> %v was not %v", lhs, rhs, expected)
	}
}

func TestPeriodCompare(t *testing.T) {
	lhs1 := testPeriod(1)
	rhs1 := testPeriod(1)
	periodCompareTest(t, &lhs1, &rhs1, comparison.SpaceshipEqual)

	lhs2 := testPeriod(1)
	rhs2 := testPeriod(2)
	periodCompareTest(t, &lhs2, &rhs2, comparison.SpaceshipLeftGreater)

	lhs3 := testPeriod(2)
	rhs3 := testPeriod(1)
	periodCompareTest(t, &lhs3, &rhs3, comparison.SpaceshipRightGreater)
}
