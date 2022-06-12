package filter

import (
	"math"

	"github.com/gotracker/gomixing/volume"
	"github.com/gotracker/voice/period"

	"github.com/gotracker/gotracker/internal/filter"
	"github.com/gotracker/gotracker/internal/optional"
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

	enabled             bool
	resonance           optional.Value[uint8]
	cutoff              optional.Value[uint8]
	playbackRate        period.Frequency
	highpass            bool
	extendedFilterRange bool
}

// NewResonantFilter creates a new resonant filter with the provided cutoff and resonance values
func NewResonantFilter(cutoff uint8, resonance uint8, playbackRate period.Frequency, extendedFilterRange bool, highpass bool) filter.Filter {
	rf := &ResonantFilter{
		playbackRate:        playbackRate,
		highpass:            highpass,
		extendedFilterRange: extendedFilterRange,
	}

	if resonance&0x80 != 0 {
		rf.resonance.Set(uint8(resonance) & 0x7f)
	}
	c := uint8(0x7F)
	if (cutoff & 0x80) != 0 {
		c = cutoff & 0x7f
		rf.cutoff.Set(uint8(c))
	}

	rf.recalculate(int8(c))
	return rf
}

func (f *ResonantFilter) Clone() filter.Filter {
	c := *f
	c.channels = make([]channelData, len(f.channels))
	for i := range f.channels {
		c.channels[i] = f.channels[i]
	}
	return &c
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

		yn := s
		if f.enabled {
			yn *= f.a0
			yn += c.ynz1*f.b0 + c.ynz2*f.b1
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
	cutoff, useCutoff := f.cutoff.Get()
	resonance, useResonance := f.resonance.Get()

	if !useResonance {
		resonance = 0
	}

	if !useCutoff {
		cutoff = 127
	} else {
		cutoff = uint8(v)
		if cutoff < 0 {
			cutoff = 0
		} else if cutoff > 127 {
			cutoff = 127
		}

		f.cutoff.Set(uint8(cutoff))
	}

	computedCutoff := int(cutoff) * 2

	useFilter := true
	if computedCutoff >= 254 && resonance == 0 {
		useFilter = false
	}

	f.enabled = useFilter
	if !f.enabled {
		return
	}

	const (
		itFilterRange  = 24.0 // standard IT range
		extfilterRange = 20.0 // extended OpenMPT range
	)

	filterRange := itFilterRange
	if f.extendedFilterRange {
		filterRange = extfilterRange
	}

	const dampingFactorDivisor = ((24.0 / 128.0) / 20.0)
	dampingFactor := math.Pow(10.0, -float64(resonance)*dampingFactorDivisor)

	f2 := float64(f.playbackRate) / 2.0
	freq := f2
	if computedCutoff < 254 {
		fcComputedCutoff := float64(computedCutoff)
		freq = 110.0 * math.Pow(2.0, 0.25+(fcComputedCutoff/filterRange))
		if freq < 120.0 {
			freq = 120.0
		} else if freq > 20000 {
			freq = 20000
		}
	}
	if freq > f2 {
		freq = f2
	}

	fc := freq * 4.0 * math.Pi

	var d, e float64
	if f.extendedFilterRange {
		r := fc / float64(f.playbackRate)

		d = (1.0 - 2.0*dampingFactor) * r
		if d > 2.0 {
			d = 2.0
		}
		d = (2.0*dampingFactor - d) / r
		e = 1.0 / (r * r)
	} else {
		r := float64(f.playbackRate) / fc

		d = dampingFactor*r + dampingFactor - 1.0
		e = r * r
	}

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
