package note

// Period is an interface that defines a sampler period
type Period interface {
	Add(PeriodDelta) Period
	Compare(Period) SpaceshipResult // <=>
	Lerp(float64, Period) Period
	GetSamplerAdd(float64) float64
	GetFrequency() float64
}

// PeriodDelta is an amount of delta specific to the period type it modifies
// it's intended to be non-specific unit type, so it's up to the implementer
// to keep track of the expected unit type.
type PeriodDelta float64

// ComparePeriods compares two periods, taking nil into account
func ComparePeriods(lhs Period, rhs Period) SpaceshipResult {
	if lhs == nil {
		if rhs == nil {
			return CompareEqual
		}
		return CompareRightHigher
	} else if rhs == nil {
		return CompareLeftHigher
	}

	return lhs.Compare(rhs)
}
