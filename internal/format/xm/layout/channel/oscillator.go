package channel

import (
	"math/rand"

	formatutil "gotracker/internal/format/internal/util"
)

// WaveTableSelect is the selection code for which waveform to use in an oscillator
type WaveTableSelect uint8

const (
	// WaveTableSelectSineRetrigger is for a sine wave that retriggers when a new note is played
	WaveTableSelectSineRetrigger = WaveTableSelect(iota)
	// WaveTableSelectSawtoothRetrigger is for a sawtooth wave that retriggers when a new note is played
	WaveTableSelectSawtoothRetrigger
	// WaveTableSelectSquareRetrigger is for a square wave that retriggers when a new note is played
	WaveTableSelectSquareRetrigger
	// WaveTableSelectRandomRetrigger is for random data wave that retriggers when a new note is played
	WaveTableSelectRandomRetrigger
	// WaveTableSelectSineContinue is for a sine wave that does not retrigger when a new note is played
	WaveTableSelectSineContinue
	// WaveTableSelectSawtoothContinue is for a sawtooth wave that does not retrigger when a new note is played
	WaveTableSelectSawtoothContinue
	// WaveTableSelectSquareContinue is for a square wave that does not retrigger when a new note is played
	WaveTableSelectSquareContinue
	// WaveTableSelectRandomContinue is for random data wave that does not retrigger when a new note is played
	WaveTableSelectRandomContinue
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
	case WaveTableSelectSineRetrigger, WaveTableSelectSineContinue:
		vib = formatutil.GetProtrackerSine(int(o.Pos))
	case WaveTableSelectSawtoothRetrigger, WaveTableSelectSawtoothContinue:
		vib = (32.0 - float32(o.Pos&0x3f)) / 32.0
	case WaveTableSelectSquareRetrigger, WaveTableSelectSquareContinue:
		v := formatutil.GetProtrackerSine(int(o.Pos))
		if v > 0 {
			vib = 1.0
		} else {
			vib = -1.0
		}
	case WaveTableSelectRandomRetrigger, WaveTableSelectRandomContinue:
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
