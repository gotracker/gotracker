package envelope

type PitchPoint struct {
	Ticks int
	Y     float32
}

func (p PitchPoint) Length() int {
	return p.Ticks
}

func (p PitchPoint) Value(out interface{}) {
	*out.(*float32) = p.Y
}

func (p *PitchPoint) Init(ticks int, value interface{}) {
	p.Ticks = ticks
	p.Y = value.(float32)
}
