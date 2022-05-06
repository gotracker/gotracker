package filter

import (
	"math"

	"github.com/gotracker/gomixing/volume"
)

type channelData struct {
	xnz1 volume.Volume
	//ynz1 volume.Volume
	xnz2 volume.Volume
	//ynz2 volume.Volume
}

// AmigaLPF is a 12dB/octave 2-pole Butterworth Low-Pass Filter with 3275 Hz cut-off
type AmigaLPF struct {
	channels []channelData
}

// NewAmigaLPF creates a new AmigaLPF
func NewAmigaLPF() *AmigaLPF {
	lpf := AmigaLPF{}

	return &lpf
}

const (
	//amigaLPFCoeffA0 volume.Volume = 1.0
	amigaLPFCoeffA1 volume.Volume = math.Sqrt2
	//amigaLPFCoeffA2 volume.Volume = 1.0

	//amigaLPFCoeffB0 volume.Volume = 1.0
	//amigaLPFCoeffB1 volume.Volume = 0.0
	//amigaLPFCoeffB2 volume.Volume = 0.0
)

// Filter processes incoming (dry) samples and produces an outgoing filtered (wet) result
func (f *AmigaLPF) Filter(dry volume.Matrix) volume.Matrix {
	if dry.Channels == 0 {
		return volume.Matrix{}
	}
	wet := dry
	for i := 0; i < dry.Channels; i++ {
		s := dry.StaticMatrix[i]
		for len(f.channels) <= i {
			f.channels = append(f.channels, channelData{})
		}
		c := &f.channels[i]
		xn := s
		//yn := amigaLPFCoeffA0*xn + amigaLPFCoeffA1*c.xnz1 + amigaLPFCoeffA2*c.xnz2 - amigaLPFCoeffB1*c.ynz1 - amigaLPFCoeffB2*c.ynz2
		yn := (xn + amigaLPFCoeffA1*c.xnz1 + c.xnz2) / 3 // since B1 and B2 are 0, they simplify out.  Similarly, the multiply on A0 and A2 become simpler.

		c.xnz2 = c.xnz1
		c.xnz1 = xn
		//c.ynz2 = c.ynz1
		//c.ynz1 = yn
		wet.StaticMatrix[i] = yn
	}
	return wet
}

// UpdateEnv updates the filter with the value from the filter envelope
func (f *AmigaLPF) UpdateEnv(v float32) {
}
