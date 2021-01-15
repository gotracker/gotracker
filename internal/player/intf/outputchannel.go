package intf

import (
	"github.com/gotracker/gomixing/volume"
)

// OutputChannel is the important bits to make output to a particular downmixing channel work
type OutputChannel struct {
	ChannelNum int
	Filter     Filter
	Playback   Playback
}

// ApplyFilter will apply the channel filter, if there is one.
func (oc *OutputChannel) ApplyFilter(dry volume.Matrix) volume.Matrix {
	if oc.Filter != nil {
		wet := oc.Filter.Filter(dry)
		return wet
	}
	return dry
}
