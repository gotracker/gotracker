package envelope

import (
	"github.com/gotracker/gomixing/panning"
)

type PanPoint struct {
	Ticks int
	Y     panning.Position
}

func (p PanPoint) Length() int {
	return p.Ticks
}

func (p PanPoint) Value(out interface{}) {
	*out.(*panning.Position) = p.Y
}

func (p *PanPoint) Init(ticks int, value interface{}) {
	p.Ticks = ticks
	p.Y = value.(panning.Position)
}
