package mixer

import (
	"gotracker/internal/player/volume"
	"math"
)

// PanMixer is a mixer that's specialized for mixing multichannel audio content
type PanMixer interface {
	GetMixingMatrix(float32) []volume.Volume
}

var (
	// PanMixerMono is a mixer that's specialized for mixing monaural audio content
	PanMixerMono PanMixer = &panMixerMono{}

	// PanMixerStereo is a mixer that's specialized for mixing stereo audio content
	PanMixerStereo PanMixer = &panMixerStereo{}
)

type panMixerMono struct {
	PanMixer
}

func (p panMixerMono) GetMixingMatrix(pan float32) []volume.Volume {
	return []volume.Volume{1.0}
}

type panMixerStereo struct {
	PanMixer
}

func (p panMixerStereo) GetMixingMatrix(pan float32) []volume.Volume {
	pangle := math.Pi * float64(pan) / 2.0
	l := volume.Volume(math.Cos(pangle))
	r := volume.Volume(math.Sin(pangle))
	return []volume.Volume{l, r}
}
