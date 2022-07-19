package panning

import "github.com/gotracker/gomixing/panning"

var (
	// DefaultPanningLeft is the default panning value for left channels
	DefaultPanningLeft = PanningFromS3M(0x03)
	// DefaultPanning is the default panning value for unconfigured channels
	DefaultPanning = PanningFromS3M(0x08)
	// DefaultPanningRight is the default panning value for right channels
	DefaultPanningRight = PanningFromS3M(0x0C)
)

// PanningFromS3M returns a radian panning position from an S3M panning value
func PanningFromS3M(pos uint8) panning.Position {
	return panning.MakeStereoPosition(float32(pos), 0, 0x0F)
}
