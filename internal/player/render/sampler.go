package render

type Sampler struct {
	SampleRate    int
	Channels      int
	BitsPerSample int
	BaseClockRate float32
}

func (s *Sampler) GetSamplerSpeed() float32 {
	return s.BaseClockRate / float32(s.SampleRate)
}
