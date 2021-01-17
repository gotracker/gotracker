package intf

import (
	"github.com/gotracker/gomixing/volume"
)

// OutputChannel is the important bits to make output to a particular downmixing channel work
type OutputChannel struct {
	ChannelNum   int
	Filter       Filter
	Playback     Playback
	PreMixVolume volume.Volume
}

// ApplyFilter will apply the channel filter, if there is one.
func (oc *OutputChannel) ApplyFilter(dry volume.Matrix) volume.Matrix {
	wet := oc.PreMixVolume.Apply(dry...)
	if oc.Filter != nil {
		wet = oc.Filter.Filter(wet)
		return wet
	}
	return wet
}
