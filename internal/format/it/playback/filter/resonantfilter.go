package filter

import (
	"math"

	"github.com/gotracker/gomixing/volume"

	"github.com/gotracker/gotracker/internal/filter"
	"github.com/gotracker/gotracker/internal/format/it/playback/util"
)

type channelData struct {
	ynz1 volume.Volume
	ynz2 volume.Volume
}

var (
	rfFreqParamMult         float64 = 128.0 / (24.0 * 256.0)
	rfPeriodResonanceFactor float64 = 2.0 * math.Pi / float64(util.C5Period)
)

// ResonantFilter is a modified 2-pole resonant filter
type ResonantFilter struct {
	channels []channelData
	a0       volume.Volume
	b0       volume.Volume
	b1       volume.Volume
	f2       float64
	rf       float64
	cm       float64
}

// NewResonantFilter creates a new resonant filter with the provided cutoff and resonance values
func NewResonantFilter(cutoff uint8, resonance uint8, playbackRate float32) filter.Filter {
	rf := &ResonantFilter{
		f2: float64(playbackRate) / 2.0,
		rf: rfPeriodResonanceFactor * float64(resonance),
		cm: float64(cutoff) * rfFreqParamMult,
	}

	rf.recalculate(255)
	return rf
}

// Filter processes incoming (dry) samples and produces an outgoing filtered (wet) result
func (f *ResonantFilter) Filter(dry volume.Matrix) volume.Matrix {
	if dry.Channels == 0 {
		return volume.Matrix{}
	}
	wet := dry // we can update in-situ and be ok
	for i := 0; i < dry.Channels; i++ {
		s := dry.StaticMatrix[i]
		for len(f.channels) <= i {
			f.channels = append(f.channels, channelData{})
		}
		c := &f.channels[i]

		xn := s
		yn := (xn*f.a0 + c.ynz1*f.b0 + c.ynz2*f.b1) / 3
		c.ynz2 = c.ynz1
		c.ynz1 = yn
		wet.StaticMatrix[i] = yn
	}
	return wet
}

func (f *ResonantFilter) recalculate(v float32) {
	co := (f.cm * float64(v+256)) / 256
	if co > 255 {
		co = 255
	}
	freq := 110.0 * math.Pow(2.0, co+0.25)
	if freq > f.f2 {
		freq = f.f2
	}
	r := f.f2 / (math.Pi * freq)

	resoFactor := 1.0 - f.rf
	d := resoFactor*r + resoFactor - 1.0
	e := r * r

	de1 := 1.0 + d + e

	fg := 1.0 / de1
	fb0 := (d + e + e) / de1
	fb1 := -e / de1

	f.a0 = volume.Volume(fg)
	f.b0 = volume.Volume(fb0)
	f.b1 = volume.Volume(fb1)
}

// UpdateEnv updates the filter with the value from the filter envelope
func (f *ResonantFilter) UpdateEnv(v float32) {
	f.recalculate(v * 255)
}
