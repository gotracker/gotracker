package channel

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
