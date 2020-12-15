package render

import (
	"gotracker/internal/player/render/mixer"
)

// Sampler is a container of sampler/mixer settings
type Sampler struct {
	SampleRate    int
	BaseClockRate float32

	mixer mixer.Mixer
}

// NewSampler returns a new sampler object based on the input settings
func NewSampler(samplesPerSec int, channels int, bitsPerSample int) *Sampler {
	s := Sampler{
		SampleRate: samplesPerSec,
		mixer: mixer.Mixer{
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
func (s *Sampler) Mixer() *mixer.Mixer {
	return &s.mixer
}

// ToRenderData converts a mixbuffer into a byte stream intended to be
// output to the output sound device
func (s *Sampler) ToRenderData(data mixer.MixBuffer, mixedChannels int) []byte {
	if len(data) == 0 {
		return nil
	}
	samples := len(data[0])
	return data.ToRenderData(samples, s.mixer.BitsPerSample, mixedChannels)
}

// GetPanMixer returns the panning mixer that can generate a matrix
// based on input pan value
func (s *Sampler) GetPanMixer() mixer.PanMixer {
	return mixer.GetPanMixer(s.mixer.Channels)
}
