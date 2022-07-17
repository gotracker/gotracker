package oscillator

import (
	"math/rand"

	"github.com/gotracker/voice/oscillator"
)

var (
	impulseSineTable = [...]int8{
		0, 2, 3, 5, 6, 8, 9, 11, 12, 14, 16, 17, 19, 20, 22, 23,
		24, 26, 27, 29, 30, 32, 33, 34, 36, 37, 38, 39, 41, 42, 43, 44,
		45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 56, 57, 58, 59,
		59, 60, 60, 61, 61, 62, 62, 62, 63, 63, 63, 64, 64, 64, 64, 64,
		64, 64, 64, 64, 64, 64, 63, 63, 63, 62, 62, 62, 61, 61, 60, 60,
		59, 59, 58, 57, 56, 56, 55, 54, 53, 52, 51, 50, 49, 48, 47, 46,
		45, 44, 43, 42, 41, 39, 38, 37, 36, 34, 33, 32, 30, 29, 27, 26,
		24, 23, 22, 20, 19, 17, 16, 14, 12, 11, 9, 8, 6, 5, 3, 2,
		0, -2, -3, -5, -6, -8, -9, -11, -12, -14, -16, -17, -19, -20, -22, -23,
		-24, -26, -27, -29, -30, -32, -33, -34, -36, -37, -38, -39, -41, -42, -43, -44,
		-45, -46, -47, -48, -49, -50, -51, -52, -53, -54, -55, -56, -56, -57, -58, -59,
		-59, -60, -60, -61, -61, -62, -62, -62, -63, -63, -63, -64, -64, -64, -64, -64,
		-64, -64, -64, -64, -64, -64, -63, -63, -63, -62, -62, -62, -61, -61, -60, -60,
		-59, -59, -58, -57, -56, -56, -55, -54, -53, -52, -51, -50, -49, -48, -47, -46,
		-45, -44, -43, -42, -41, -39, -38, -37, -36, -34, -33, -32, -30, -29, -27, -26,
		-24, -23, -22, -20, -19, -17, -16, -14, -12, -11, -9, -8, -6, -5, -3, -2,
	}

	impulseSawtoothTable = [...]int8{
		64, 63, 63, 62, 62, 61, 61, 60, 60, 59, 59, 58, 58, 57, 57, 56,
		56, 55, 55, 54, 54, 53, 53, 52, 52, 51, 51, 50, 50, 49, 49, 48,
		48, 47, 47, 46, 46, 45, 45, 44, 44, 43, 43, 42, 42, 41, 41, 40,
		40, 39, 39, 38, 38, 37, 37, 36, 36, 35, 35, 34, 34, 33, 33, 32,
		32, 31, 31, 30, 30, 29, 29, 28, 28, 27, 27, 26, 26, 25, 25, 24,
		24, 23, 23, 22, 22, 21, 21, 20, 20, 19, 19, 18, 18, 17, 17, 16,
		16, 15, 15, 14, 14, 13, 13, 12, 12, 11, 11, 10, 10, 9, 9, 8,
		8, 7, 7, 6, 6, 5, 5, 4, 4, 3, 3, 2, 2, 1, 1, 0,
		0, -1, -1, -2, -2, -3, -3, -4, -4, -5, -5, -6, -6, -7, -7, -8,
		-8, -9, -9, -10, -10, -11, -11, -12, -12, -13, -13, -14, -14, -15, -15, -16,
		-16, -17, -17, -18, -18, -19, -19, -20, -20, -21, -21, -22, -22, -23, -23, -24,
		-24, -25, -25, -26, -26, -27, -27, -28, -28, -29, -29, -30, -30, -31, -31, -32,
		-32, -33, -33, -34, -34, -35, -35, -36, -36, -37, -37, -38, -38, -39, -39, -40,
		-40, -41, -41, -42, -42, -43, -43, -44, -44, -45, -45, -46, -46, -47, -47, -48,
		-48, -49, -49, -50, -50, -51, -51, -52, -52, -53, -53, -54, -54, -55, -55, -56,
		-56, -57, -57, -58, -58, -59, -59, -60, -60, -61, -61, -62, -62, -63, -63, -64,
	}
)

// GetImpulseSine returns the sine value for a particular position using the
// ImpulseTracker-compliant full-period sine table
func GetImpulseSine(pos int) float32 {
	return float32(impulseSineTable[pos&0xff]) / 64
}

// GetImpulseSawtooth returns the sawtooth value for a particular position using the
// ImpulseTracker-compliant full-period sawtooth table
func GetImpulseSawtooth(pos int) float32 {
	return float32(impulseSawtoothTable[pos&0xff]) / 64
}

// impulseOscillator is an oscillator using the protracker sine table
type impulseOscillator struct {
	Table oscillator.WaveTableSelect
	Pos   uint8
	Mul   uint8
}

// NewImpulseTrackerOscillator creates a new ImpulseTracker-compatible oscillator
func NewImpulseTrackerOscillator(mul uint8) oscillator.Oscillator {
	return &impulseOscillator{
		Mul: mul,
	}
}

// GetWave returns the wave amplitude for the current position
func (o *impulseOscillator) GetWave(depth float32) float32 {
	var vib float32
	switch o.Table {
	case WaveTableSelectSineRetrigger, WaveTableSelectSineContinue:
		vib = GetImpulseSine(int(o.Pos))
	case WaveTableSelectSawtoothRetrigger, WaveTableSelectSawtoothContinue:
		vib = GetImpulseSawtooth(int(o.Pos))
	case WaveTableSelectInverseSawtoothRetrigger:
		vib = -GetImpulseSawtooth(int(o.Pos))
	case WaveTableSelectSquareRetrigger, WaveTableSelectSquareContinue:
		v := GetImpulseSine(int(o.Pos))
		if v > 0 {
			vib = 1.0
		} else {
			vib = -1.0
		}
	case WaveTableSelectRandomRetrigger, WaveTableSelectRandomContinue:
		vib = GetImpulseSine(rand.Intn(0xff))
	}
	delta := vib * depth
	return delta
}

// Advance advances the oscillator position by the specified amount
func (o *impulseOscillator) Advance(speed int) {
	o.Pos += uint8(speed) * o.Mul
}

// SetWaveform sets the waveform for the current oscillator
func (o *impulseOscillator) SetWaveform(table oscillator.WaveTableSelect) {
	o.Table = table
}

// Reset resets the position of the oscillator
func (o *impulseOscillator) Reset(hard ...bool) {
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
