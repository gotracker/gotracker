package intf

import (
	"gotracker/internal/song"

	"github.com/gotracker/gomixing/volume"
)

// OutputChannel is the important bits to make output to a particular downmixing channel work
type OutputChannel struct {
	ChannelNum    int
	Filter        song.Filter
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

// SetFilterEnvelopeValue updates the filter on the channel with the new envelope value
func (oc *OutputChannel) SetFilterEnvelopeValue(envVal float32) {
	if oc.Filter != nil {
		oc.Filter.UpdateEnv(envVal)
	}
}
