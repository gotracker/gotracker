package panning

import (
	itfile "github.com/gotracker/goaudiofile/music/tracked/it"
	"github.com/gotracker/gomixing/panning"
)

var (
	// DefaultPanningLeft is the default panning value for left channels
	DefaultPanningLeft = FromItPanning(0x30)
	// DefaultPanning is the default panning value for unconfigured channels
	DefaultPanning = FromItPanning(0x80)
	// DefaultPanningRight is the default panning value for right channels
	DefaultPanningRight = FromItPanning(0xC0)
)

// FromItPanning returns a radian panning position from an it panning value
func FromItPanning(pos itfile.PanValue) panning.Position {
	if pos.IsDisabled() {
		return panning.CenterAhead
	}
	return panning.MakeStereoPosition(pos.Value(), 0, 1)
}

// ToItPanning returns the it panning value for a radian panning position
func ToItPanning(pan panning.Position) itfile.PanValue {
	p := panning.FromStereoPosition(pan, 0, 1)
	return itfile.PanValue(p * 64)
}
