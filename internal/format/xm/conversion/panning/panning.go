package panning

import (
	"github.com/gotracker/gomixing/panning"
)

var (
	// DefaultPanningLeft is the default panning value for left channels
	DefaultPanningLeft = PanningFromXm(0x30)
	// DefaultPanning is the default panning value for unconfigured channels
	DefaultPanning = PanningFromXm(0x80)
	// DefaultPanningRight is the default panning value for right channels
	DefaultPanningRight = PanningFromXm(0xC0)
)

// PanningFromXm returns a radian panning position from an xm panning value
func PanningFromXm(pos uint8) panning.Position {
	return panning.MakeStereoPosition(float32(pos), 0, 0xFF)
}

// PanningToXm returns the xm panning value for a radian panning position
func PanningToXm(pan panning.Position) uint8 {
	return uint8(panning.FromStereoPosition(pan, 0, 0xFF))
}
