package filter

import (
	"math"

	"github.com/gotracker/gomixing/volume"

	"github.com/gotracker/gotracker/internal/filter"
)

type channelData struct {
	ynz1 volume.Volume
	ynz2 volume.Volume
}

// ResonantFilter is a modified 2-pole resonant filter
type ResonantFilter struct {
	channels []channelData
	a0       volume.Volume
	b0       volume.Volume
	b1       volume.Volume

	enabled      bool
	resonance    int
	cutoff       uint8
	playbackRate int
	filterRange  float64
	highpass     bool
}

// NewResonantFilter creates a new resonant filter with the provided cutoff and resonance values
func NewResonantFilter(cutoff uint8, resonance uint8, playbackRate int, extendedFilterRange bool, highpass bool) filter.Filter {
	r := resonance
	if r&0x80 != 0 {
		r = 0
	}
	c := cutoff
	if (c & 0x80) != 0 {
		c = 0x7F
	}
	const itFilterRange = 24.0  // standard IT range
	const extfilterRange = 20.0 // extended OpenMPT range
	rf := &ResonantFilter{
		cutoff:       c,
		resonance:    int(r),
		playbackRate: playbackRate,
		filterRange:  itFilterRange,
		highpass:     highpass,
	}

	if extendedFilterRange {
		rf.filterRange = extfilterRange
	}

	rf.recalculate(int8(c))
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
		yn := xn
		if f.enabled {
			yn = (xn*f.a0 + c.ynz1*f.b0 + c.ynz2*f.b1)
		}
		c.ynz2 = c.ynz1
		c.ynz1 = yn
		if f.highpass {
			c.ynz1 -= s
		}
		wet.StaticMatrix[i] = yn
	}
	return wet
}

func (f *ResonantFilter) recalculate(v int8) {
	cutoff := int(v)
	resonance := f.resonance

	if cutoff < 0 {
		cutoff = 0
	} else if cutoff > 127 {
		cutoff = 127
	}

	if resonance < 0 {
		resonance = 0
	} else if resonance > 127 {
		resonance = 127
	}

	f.cutoff = uint8(cutoff)
	f.resonance = resonance

	computedCutoff := int(cutoff) * 2

	if resonance == 0 || computedCutoff >= 254 {
		f.enabled = false
		return
	}

	f.enabled = true

	const dampingFactorDivisor = ((24.0 / 128.0) / 20.0)
	dampingFactor := math.Pow(10.0, -float64(resonance)*dampingFactorDivisor)

	fcComputedCutoff := float64(computedCutoff)
	freq := 110.0 * math.Pow(2.0, 0.25+fcComputedCutoff/f.filterRange)
	if freq < 120.0 {
		freq = 120.0
	} else if freq > 20000 {
		freq = 20000
	}
	f2 := float64(f.playbackRate) / 2.0
	if freq > f2 {
		freq = f2
	}

	fc := freq * 2.0 * math.Pi

	r := float64(f.playbackRate) / fc

	d := dampingFactor*r + dampingFactor - 1.0
	e := r * r

	a := 1.0 / (1.0 + d + e)
	b := (d + e + e) * a
	c := -e * a
	if f.highpass {
		a = 1.0 - a
	} else {
		// lowpass
		if a == 0 {
			// prevent silence at extremely low cutoff and very high sampling rate
			a = 1.0
		}
	}

	f.a0 = volume.Volume(a)
	f.b0 = volume.Volume(b)
	f.b1 = volume.Volume(c)
}

// UpdateEnv updates the filter with the value from the filter envelope
func (f *ResonantFilter) UpdateEnv(cutoff int8) {
	f.recalculate(cutoff)
}
