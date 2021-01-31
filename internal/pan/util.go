package pan

import (
	"math"

	"github.com/gotracker/gomixing/panning"
)

const (
	pi2 = math.Pi / 2
	pi4 = math.Pi / 4
	pi8 = math.Pi / 8

	twopi = math.Pi * 2
)

// CalculateCombinedPanning calculates a panning value where `p1` modifies
// panning value `p0` such that `p0` is primary component and `p1` is secondary
// TODO: JBC - move this calculation function into gomixing lib
func CalculateCombinedPanning(p0, p1 panning.Position) panning.Position {
	p0a := float64(p0.Angle)
	p1a := float64(p1.Angle)

	fa := p0a + (p1a-pi8)*(pi4-math.Abs(p0a-pi4))/pi8
	if fa > pi2 {
		fa = pi2
	} else if fa < 0 {
		fa = 0
	}

	fd := math.Sqrt(float64(p0.Distance * p1.Distance))

	return panning.Position{
		Angle:    float32(fa),
		Distance: float32(fd),
	}
}

// GetPanningDifference calculates the difference of `p0` - `p1`
func GetPanningDifference(p0, p1 panning.Position) panning.Position {
	ia := float64(panning.CenterAhead.Angle)
	p0a := float64(p0.Angle)
	p1a := float64(p1.Angle)

	fa := math.Mod(ia+p1a-p0a, twopi)
	for fa < 0 {
		fa += twopi
	}
	fd := panning.CenterAhead.Distance + p1.Distance - p0.Distance

	return panning.Position{
		Angle:    float32(fa),
		Distance: float32(fd),
	}
}
