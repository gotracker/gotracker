package panning

// Position is stored as polar coordinates
// with Angle of 0 radians being calculated from right
// and >0 rotating counter-clockwise from that point
type Position struct {
	Angle    float32
	Distance float32
}
