package oscillator

import (
	"math/rand"

	"github.com/gotracker/voice/oscillator"
)

var (
	protrackerSineTable = [32]uint8{
		0, 24, 49, 74, 97, 120, 141, 161, 180, 197, 212, 224, 235, 244, 250, 253,
		255, 253, 250, 244, 235, 224, 212, 197, 180, 161, 141, 120, 97, 74, 49, 24,
	}
)

// GetProtrackerSine returns the sine value for a particular position using the
// Protracker-compliant half-period sine table
func GetProtrackerSine(pos int) float32 {
	sin := float32(protrackerSineTable[pos&0x1f]) / 255
	if pos > 32 {
		return -sin
	}
	return sin
}

// protrackerOscillator is an oscillator using the protracker sine table
type protrackerOscillator struct {
	Table oscillator.WaveTableSelect
	Pos   int8
}

// NewProtrackerOscillator creates a new Protracker-compatible oscillator
func NewProtrackerOscillator() oscillator.Oscillator {
	return &protrackerOscillator{}
}

// GetWave returns the wave amplitude for the current position
func (o *protrackerOscillator) GetWave(depth float32) float32 {
	var vib float32
	switch o.Table {
	case WaveTableSelectSineRetrigger, WaveTableSelectSineContinue:
		vib = GetProtrackerSine(int(o.Pos))
	case WaveTableSelectSawtoothRetrigger, WaveTableSelectSawtoothContinue:
		vib = (32.0 - float32(o.Pos&0x3f)) / 32.0
	case WaveTableSelectInverseSawtoothRetrigger:
		vib = -(32.0 - float32(o.Pos&0x3f)) / 32.0
	case WaveTableSelectSquareRetrigger, WaveTableSelectSquareContinue:
		v := GetProtrackerSine(int(o.Pos))
		if v > 0 {
			vib = 1.0
		} else {
			vib = -1.0
		}
	case WaveTableSelectRandomRetrigger, WaveTableSelectRandomContinue:
		vib = GetProtrackerSine(rand.Intn(0x3f))
	}
	delta := vib * depth
	return delta
}

// Advance advances the oscillator position by the specified amount
func (o *protrackerOscillator) Advance(speed int) {
	o.Pos += int8(speed)
	for o.Pos < 0 {
		o.Pos += 64
	}
	for o.Pos > 63 {
		o.Pos -= 64
	}
}

// SetWaveform sets the waveform for the current oscillator
func (o *protrackerOscillator) SetWaveform(table oscillator.WaveTableSelect) {
	o.Table = table
}

// Reset resets the position of the oscillator
func (o *protrackerOscillator) Reset(hard ...bool) {
	hardReset := false
	if len(hard) > 0 {
		hardReset = hard[0]
	}

	doReset := hardReset
	switch o.Table {
	case WaveTableSelectSineRetrigger, WaveTableSelectSawtoothRetrigger, WaveTableSelectSquareRetrigger, WaveTableSelectRandomRetrigger:
		doReset = true
	}

	if doReset {
		o.Pos = 0
	}
}
