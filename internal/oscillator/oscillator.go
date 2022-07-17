package oscillator

import (
	"github.com/gotracker/voice/oscillator"
)

const (
	// WaveTableSelectSineRetrigger is for a sine wave that retriggers when a new note is played
	WaveTableSelectSineRetrigger = oscillator.WaveTableSelect(iota)
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
	// WaveTableSelectInverseSawtoothRetrigger is for a sawtooth wave that retriggers when a new note is played and has negated amplitude
	WaveTableSelectInverseSawtoothRetrigger
)
