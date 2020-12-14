package mixer

import (
	"gotracker/internal/player/panning"
	"gotracker/internal/player/volume"
	"math"
)

// PanMixer is a mixer that's specialized for mixing multichannel audio content
type PanMixer interface {
	GetMixingMatrix(panning.Position) volume.VolumeMatrix
}

var (
	// PanMixerMono is a mixer that's specialized for mixing monaural audio content
	PanMixerMono PanMixer = &panMixerMono{}

	// PanMixerStereo is a mixer that's specialized for mixing stereo audio content
	PanMixerStereo PanMixer = &panMixerStereo{}

	// PanMixerQuad is a mixer that's specialized for mixing quadraphonic audio content
	PanMixerQuad PanMixer = &panMixerQuad{}
)

type panMixerMono struct {
	PanMixer
}

func (p panMixerMono) GetMixingMatrix(pan panning.Position) volume.VolumeMatrix {
	// distance and angle are ignored on mono
	return volume.VolumeMatrix{1.0}
}

type panMixerStereo struct {
	PanMixer
}

func (p panMixerStereo) GetMixingMatrix(pan panning.Position) volume.VolumeMatrix {
	pangle := float64(pan.Angle)
	d := volume.Volume(pan.Distance)
	l := d * volume.Volume(math.Cos(pangle))
	r := d * volume.Volume(math.Sin(pangle))
	return volume.VolumeMatrix{l, r}
}

type panMixerQuad struct {
	PanMixer
}

func (p panMixerQuad) GetMixingMatrix(pan panning.Position) volume.VolumeMatrix {
	pangle := float64(pan.Angle)
	d := volume.Volume(pan.Distance)
	lf := d * volume.Volume(math.Cos(pangle))
	rf := d * volume.Volume(math.Sin(pangle))
	lr := d * volume.Volume(math.Sin(pangle+math.Pi/2.0))
	rr := d * volume.Volume(math.Cos(pangle-math.Pi/2.0))
	return volume.VolumeMatrix{lf, rf, lr, rr}
}

// GetPanMixer returns the panning mixer that can generate a matrix
// based on input pan value
func GetPanMixer(channels int) PanMixer {
	switch channels {
	case 1:
		return PanMixerMono
	case 2:
		return PanMixerStereo
	case 4:
		return PanMixerQuad
	}

	return nil
}
