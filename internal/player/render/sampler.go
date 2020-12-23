package render

import (
	"github.com/heucuva/gomixing/mixing"
)

// Sampler is a container of sampler/mixer settings
type Sampler struct {
	SampleRate    int
	BaseClockRate float32

	mixer mixing.Mixer
}

// NewSampler returns a new sampler object based on the input settings
func NewSampler(samplesPerSec int, channels int, bitsPerSample int) *Sampler {
	s := Sampler{
		SampleRate: samplesPerSec,
		mixer: mixing.Mixer{
			Channels:      channels,
			BitsPerSample: bitsPerSample,
		},
	}
	return &s
}

// GetSamplerSpeed returns the current sampler speed
// which is a product of the base sampler clock rate and the inverse
// of the output render rate (the sample rate)
func (s *Sampler) GetSamplerSpeed() float32 {
	return s.BaseClockRate / float32(s.SampleRate)
}

// Mixer returns a pointer to the current mixer object
func (s *Sampler) Mixer() *mixing.Mixer {
	return &s.mixer
}

// GetPanMixer returns the panning mixer that can generate a matrix
// based on input pan value
func (s *Sampler) GetPanMixer() mixing.PanMixer {
	return mixing.GetPanMixer(s.mixer.Channels)
}
