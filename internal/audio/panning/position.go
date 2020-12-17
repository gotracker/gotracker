package panning

import "math"

// Position is stored as polar coordinates
// with Angle of 0 radians being calculated from right
// and >0 rotating counter-clockwise from that point
type Position struct {
	Angle    float32
	Distance float32
}

var (
	// CenterAhead is the position directly ahead of the listener
	CenterAhead = Position{
		Angle:    float32(math.Pi / 2.0),
		Distance: float32(1.0),
	}
)
