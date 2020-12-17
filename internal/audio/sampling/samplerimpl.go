package sampling

import "gotracker/internal/audio/volume"

type sampler struct {
	Sampler
	ss     SampleStream
	pos    Pos
	period float32
}

func (s *sampler) GetPosition() Pos {
	return s.pos
}

func (s *sampler) Advance() {
	s.pos.Add(s.period)
}

func (s *sampler) GetSample() volume.VolumeMatrix {
	if s.ss == nil {
		return nil
	}
	return s.ss.GetSample(s.pos)
}
