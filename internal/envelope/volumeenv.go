package envelope

import "github.com/gotracker/gomixing/volume"

type VolumePoint struct {
	Ticks int
	Y     volume.Volume
}

func (p VolumePoint) Length() int {
	return p.Ticks
}

func (p VolumePoint) Value(out interface{}) {
	*out.(*volume.Volume) = p.Y
}

func (p *VolumePoint) Init(ticks int, value interface{}) {
	p.Ticks = ticks
	p.Y = value.(volume.Volume)
}
