package channel

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
