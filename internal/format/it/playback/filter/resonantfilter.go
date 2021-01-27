package filter

import (
	"math"

	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/format/it/playback/util"
	"gotracker/internal/player/intf"
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
}

// NewResonantFilter creates a new resonant filter with the provided cutoff and resonance values
func NewResonantFilter(cutoff uint8, resonance uint8, mixingRate float32) intf.Filter {
	rf := ResonantFilter{}

	freq := 110.0 * math.Pow(2.0, float64(cutoff)*rfFreqParamMult+0.25)
	f2 := float64(mixingRate) / 2.0
	if freq > f2 {
		freq = f2
	}
	r := f2 / (math.Pi * freq)

	resoFactor := 1.0 - rfPeriodResonanceFactor*float64(resonance)
	d := resoFactor*r + resoFactor - 1.0
	e := r * r

	de1 := 1.0 + d + e

	fg := 1.0 / de1
	fb0 := (d + e + e) / de1
	fb1 := -e / de1

	rf.a0 = volume.Volume(fg)
	rf.b0 = volume.Volume(fb0)
	rf.b1 = volume.Volume(fb1)

	return &rf
}

// Filter processes incoming (dry) samples and produces an outgoing filtered (wet) result
func (f *ResonantFilter) Filter(dry volume.Matrix) volume.Matrix {
	wet := dry // we can update in-situ and be ok
	for i, s := range dry {
		for len(f.channels) <= i {
			f.channels = append(f.channels, channelData{})
		}
		c := &f.channels[i]

		xn := s
		yn := (xn*f.a0 + c.ynz1*f.b0 + c.ynz2*f.b1) / 3
		c.ynz2 = c.ynz1
		c.ynz1 = yn
		wet[i] = yn
	}
	return wet
}
