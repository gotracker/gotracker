package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
	"gotracker/internal/player/oscillator"
)

// SetVibratoWaveform defines a set vibrato waveform effect
type SetVibratoWaveform uint8 // 'E4x'

// PreStart triggers when the effect enters onto the channel state
func (e SetVibratoWaveform) PreStart(cs intf.Channel, p intf.Playback) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetVibratoWaveform) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	vib := cs.GetVibratoOscillator()
	vib.Table = oscillator.WaveTableSelect(x)
}

// Tick is called on every tick
func (e SetVibratoWaveform) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e SetVibratoWaveform) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e SetVibratoWaveform) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}