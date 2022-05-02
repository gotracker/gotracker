package intf

import (
	"gotracker/internal/filter"

	"github.com/gotracker/gomixing/volume"
)

// OutputChannel is the important bits to make output to a particular downmixing channel work
type OutputChannel[TChannelData any] struct {
	ChannelNum    int
	Filter        filter.Filter
	Playback      Playback
	GlobalVolume  volume.Volume
	ChannelVolume volume.Volume
}

// ApplyFilter will apply the channel filter, if there is one.
func (oc *OutputChannel[TChannelData]) ApplyFilter(dry volume.Matrix) volume.Matrix {
	premix := oc.GetPremixVolume()
	wet := dry.ApplyInSitu(premix)
	if oc.Filter != nil {
		wet = oc.Filter.Filter(wet)
		return wet
	}
	return wet
}

// GetPremixVolume returns the premix volume of the output channel
func (oc *OutputChannel[TChannelData]) GetPremixVolume() volume.Volume {
	return oc.GlobalVolume * oc.ChannelVolume
}

// SetFilterEnvelopeValue updates the filter on the channel with the new envelope value
func (oc *OutputChannel[TChannelData]) SetFilterEnvelopeValue(envVal float32) {
	if oc.Filter != nil {
		oc.Filter.UpdateEnv(envVal)
	}
}
