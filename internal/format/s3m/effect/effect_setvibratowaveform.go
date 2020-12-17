package effect

import (
	"fmt"
	"gotracker/internal/module/player/intf"
	"gotracker/internal/module/player/oscillator"
)

// SetVibratoWaveform defines a set vibrato waveform effect
type SetVibratoWaveform uint8 // 'S3x'

// PreStart triggers when the effect enters onto the channel state
func (e SetVibratoWaveform) PreStart(cs intf.Channel, ss intf.Song) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetVibratoWaveform) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	vib := cs.GetVibratoOscillator()
	vib.Table = oscillator.WaveTableSelect(x)
}

// Tick is called on every tick
func (e SetVibratoWaveform) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e SetVibratoWaveform) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e SetVibratoWaveform) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
