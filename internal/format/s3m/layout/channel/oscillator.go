package channel

import (
	"math/rand"

	formatutil "gotracker/internal/format/internal/util"
)

// WaveTableSelect is the selection code for which waveform to use in an oscillator
type WaveTableSelect uint8

const (
	// WaveTableSelectSine is for a sine wave
	WaveTableSelectSine = WaveTableSelect(iota)
	// WaveTableSelectSawtooth is for a sawtooth wave
	WaveTableSelectSawtooth
	// WaveTableSelectSquare is for a square wave
	WaveTableSelectSquare
	// WaveTableSelectRandom is for random data wave
	WaveTableSelectRandom
)

// Oscillator is an oscillator
type Oscillator struct {
	Table WaveTableSelect
	Pos   int8
}

// GetWave returns the wave amplitude for the current position
func (o *Oscillator) GetWave(depth float32) float32 {
	var vib float32
	switch o.Table {
	case WaveTableSelectSine:
		vib = formatutil.GetProtrackerSine(int(o.Pos))
	case WaveTableSelectSawtooth:
		vib = (32.0 - float32(o.Pos&0x3f)) / 32.0
	case WaveTableSelectSquare:
		v := formatutil.GetProtrackerSine(int(o.Pos))
		if v > 0 {
			vib = 1.0
		} else {
			vib = -1.0
		}
	case WaveTableSelectRandom:
		vib = formatutil.GetProtrackerSine(rand.Int() & 0x3f)
	}
	delta := vib * depth
	return delta
}

// Advance advances the oscillator position by the specified amount
func (o *Oscillator) Advance(speed int) {
	o.Pos += int8(speed)
	for o.Pos < 0 {
		o.Pos += 64
	}
	for o.Pos > 63 {
		o.Pos -= 64
	}
}
