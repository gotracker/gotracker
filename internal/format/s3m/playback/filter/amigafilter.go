package filter

import (
	"math"

	"github.com/heucuva/gomixing/volume"
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

var (
	amigaLPFCoeffA0 volume.Volume = 1.0
	amigaLPFCoeffA1 volume.Volume = volume.Volume(math.Sqrt(2.0))
	amigaLPFCoeffA2 volume.Volume = 1.0
	amigaLPFDen     volume.Volume = amigaLPFCoeffA0 + amigaLPFCoeffA1 + amigaLPFCoeffA1

	amigaLPFCoeffB0 volume.Volume = 1.0
	amigaLPFCoeffB1 volume.Volume = 0.0
	amigaLPFCoeffB2 volume.Volume = 0.0
)

// Filter processes incoming (dry) samples and produces an outgoing filtered (wet) result
func (f *AmigaLPF) Filter(dry volume.Matrix) volume.Matrix {
	wet := dry // we can update in-situ and be ok
	for i, s := range dry {
		for len(f.channels) <= i {
			f.channels = append(f.channels, channelData{})
		}
		c := &f.channels[i]
		xn := s
		//yn := amigaLPFCoeffA0*xn + amigaLPFCoeffA1*c.xnz1 + amigaLPFCoeffA2*c.xnz2 - amigaLPFCoeffB1*c.ynz1 - amigaLPFCoeffB2*c.ynz2
		yn := (xn + amigaLPFCoeffA1*c.xnz1 + c.xnz2) / amigaLPFDen // since B1 and B2 are 0, they simplify out.  Similarly, the multiply on A0 and A2 become simpler.

		c.xnz2 = c.xnz1
		c.xnz1 = xn
		//c.ynz2 = c.ynz1
		//c.ynz1 = yn
		wet[i] = yn
	}
	return wet
}
