package sampler

import (
	"github.com/gotracker/gomixing/mixing"
	"github.com/gotracker/voice/period"
)

// Sampler is a container of sampler/mixer settings
type Sampler struct {
	SampleRate    int
	BaseClockRate period.Frequency

	mixer mixing.Mixer
}

// NewSampler returns a new sampler object based on the input settings
func NewSampler(samplesPerSec int, channels int, bitsPerSample int, baseClockRate period.Frequency) *Sampler {
	s := Sampler{
		SampleRate:    samplesPerSec,
		BaseClockRate: baseClockRate,
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
	return float32(s.BaseClockRate) / float32(s.SampleRate)
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
