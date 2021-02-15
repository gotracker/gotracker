package voice

import (
	"github.com/gotracker/gomixing/sampling"
)

// Positioner is the instrument position (timeline) control interface
type Positioner interface {
	SetPos(pos sampling.Pos)
	GetPos() sampling.Pos
}
