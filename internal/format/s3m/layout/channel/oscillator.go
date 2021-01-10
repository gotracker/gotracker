package channel

import (
	"math/rand"
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

var (
	protrackerSineTable = [...]uint8{
		0, 24, 49, 74, 97, 120, 141, 161,
		180, 197, 212, 224, 235, 244, 250, 253,
		255, 253, 250, 244, 235, 224, 212, 197,
		180, 161, 141, 120, 97, 74, 49, 24,
	}
)

func getProtrackerSine(pos int) float32 {
	sin := float32(protrackerSineTable[pos&0x1f]) / 255
	if pos > 32 {
		return -sin
	}
	return sin
}

// GetWave returns the wave amplitude for the current position
func (o *Oscillator) GetWave(depth float32) float32 {
	var vib float32
	switch o.Table {
	case WaveTableSelectSine:
		vib = getProtrackerSine(int(o.Pos))
	case WaveTableSelectSawtooth:
		vib = (32.0 - float32(o.Pos&64)) / 32.0
	case WaveTableSelectSquare:
		v := getProtrackerSine(int(o.Pos))
		if v > 0 {
			vib = 1.0
		} else {
			vib = -1.0
		}
	case WaveTableSelectRandom:
		vib = getProtrackerSine(rand.Int() & 0x3f)
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
