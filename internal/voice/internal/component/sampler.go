package component

import (
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/loop"
	"gotracker/internal/pcm"
)

// Sampler is a sampler component
type Sampler struct {
	sample      pcm.Sample
	pos         sampling.Pos
	keyOn       bool
	wholeLoop   loop.Loop
	sustainLoop loop.Loop
}

// Setup sets up the sampler
func (s *Sampler) Setup(sample pcm.Sample, wholeLoop loop.Loop, sustainLoop loop.Loop) {
	s.sample = sample
	s.wholeLoop = wholeLoop
	s.sustainLoop = sustainLoop
}

// SetPos sets the current position of the sampler in the pcm data (and loops)
func (s *Sampler) SetPos(pos sampling.Pos) {
	s.pos = pos
}

// GetPos returns the current position of the sampler in the pcm data (and loops)
func (s Sampler) GetPos() sampling.Pos {
	return s.pos
}

// SetKeyOn sets the key-on value (for loop processing)
func (s *Sampler) SetKeyOn(on bool) {
	s.keyOn = on
}

// GetSample returns a multi-channel sample at the specified position
func (s Sampler) GetSample(pos sampling.Pos) volume.Matrix {
	v0 := s.getConvertedSample(pos.Pos)
	if len(v0) == 0 && ((s.keyOn && s.sustainLoop.Enabled()) || s.wholeLoop.Enabled()) {
		v01 := s.getConvertedSample(pos.Pos)
		panic(v01)
	}
	if pos.Frac == 0 {
		return v0
	}
	v1 := s.getConvertedSample(pos.Pos + 1)
	for c, s := range v1 {
		v0[c] += volume.Volume(pos.Frac) * (s - v0[c])
	}
	return v0
}

func (s Sampler) getConvertedSample(pos int) volume.Matrix {
	if s.sample == nil {
		return volume.Matrix{}
	}
	sl := s.sample.Length()
	pos, _ = loop.CalcLoopPos(s.wholeLoop, s.sustainLoop, pos, sl, s.keyOn)
	if pos < 0 || pos >= sl {
		return volume.Matrix{}
	}
	s.sample.Seek(pos)
	data, err := s.sample.Read()
	if err != nil {
		return volume.Matrix{}
	}
	return data
}
