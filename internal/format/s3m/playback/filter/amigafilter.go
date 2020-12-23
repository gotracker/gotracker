package filter

import (
	"math"

	"github.com/heucuva/gomixing/volume"
)

// AmigaLPF is a 12dB/octave 2-pole Butterworth Low-Pass Filter with 3275 Hz cut-off
type AmigaLPF struct {
	xnz1 float64
	//ynz1 float64
	xnz2 float64
	//ynz2 float64
}

// NewAmigaLPF creates a new AmigaLPF
func NewAmigaLPF() *AmigaLPF {
	lpf := AmigaLPF{}

	return &lpf
}

var (
	amigaLPFCoeffA0 = 1.0
	amigaLPFCoeffA1 = math.Sqrt(2.0)
	amigaLPFCoeffA2 = 1.0
	amigaLPFDen     = amigaLPFCoeffA0 + amigaLPFCoeffA1 + amigaLPFCoeffA1

	amigaLPFCoeffB0 = 1.0
	amigaLPFCoeffB1 = 0.0
	amigaLPFCoeffB2 = 0.0
)

// Filter processes an incoming sample and produces an outgoing filtered sample
func (f *AmigaLPF) Filter(sample volume.Volume) volume.Volume {
	xn := float64(sample)
	//yn := amigaLPFCoeffA0*xn + amigaLPFCoeffA1*f.xnz1 + amigaLPFCoeffA2*f.xnz2 - amigaLPFCoeffB1*f.ynz1 - amigaLPFCoeffB2*f.ynz2
	yn := xn + amigaLPFCoeffA1*f.xnz1 + f.xnz2 // since B1 and B2 are 0, they simplify out.  Similarly, the multiply on A0 and A2 become simpler.

	f.xnz2 = f.xnz1
	f.xnz1 = xn
	//f.ynz2 = f.ynz1
	//f.ynz1 = yn
	return volume.Volume(yn / amigaLPFDen)
}
