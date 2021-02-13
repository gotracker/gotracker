package intf

import (
	"github.com/gotracker/gomixing/volume"
)

// OutputChannel is the important bits to make output to a particular downmixing channel work
type OutputChannel struct {
	ChannelNum    int
	Filter        Filter
	Playback      Playback
	GlobalVolume  volume.Volume
	ChannelVolume volume.Volume
}

// ApplyFilter will apply the channel filter, if there is one.
func (oc *OutputChannel) ApplyFilter(dry volume.Matrix) volume.Matrix {
	premix := oc.GetPremixVolume()
	wet := dry.ApplyInSitu(premix)
	if oc.Filter != nil {
		wet = oc.Filter.Filter(wet)
		return wet
	}
	return wet
}

// GetPremixVolume returns the premix volume of the output channel
func (oc *OutputChannel) GetPremixVolume() volume.Volume {
	return oc.GlobalVolume * oc.ChannelVolume
}
