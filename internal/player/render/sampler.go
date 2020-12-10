package render

import (
	"bytes"
	"encoding/binary"
	"gotracker/internal/player/render/mixer"
	"gotracker/internal/player/volume"
)

// Sampler is a container of sampler/mixer settings
type Sampler struct {
	SampleRate    int
	Channels      int
	BitsPerSample int
	BaseClockRate float32

	mixer mixer.Mixer
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
func (s *Sampler) ToRenderData(data mixer.MixBuffer) []byte {
	if len(data) == 0 {
		return nil
	}
	samples := len(data[0])
	bps := int(s.BitsPerSample / 8)
	writer := &bytes.Buffer{}
	samplePostMultiply := volume.Volume(0.25)
	for i := 0; i < samples; i++ {
		for c := 0; c < s.Channels; c++ {
			v := data[c][i] * samplePostMultiply
			if bps == 1 {
				val := uint8(v.ToSample(s.BitsPerSample))
				binary.Write(writer, binary.LittleEndian, val)
			} else {
				val := uint16(v.ToSample(s.BitsPerSample))
				binary.Write(writer, binary.LittleEndian, val)
			}
		}
	}
	return writer.Bytes()
}
