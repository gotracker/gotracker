package voice

import (
	"github.com/gotracker/gomixing/panning"
)

// PanModulator is the instrument pan (spatial) control interface
type PanModulator interface {
	SetPan(vol panning.Position)
	GetPan() panning.Position
	GetFinalPan() panning.Position
}
