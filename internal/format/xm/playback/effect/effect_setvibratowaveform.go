package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/oscillator"
	"gotracker/internal/player/intf"
)

// SetVibratoWaveform defines a set vibrato waveform effect
type SetVibratoWaveform uint8 // 'E4x'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetVibratoWaveform) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	mem := cs.GetMemory().(*channel.Memory)
	vib := mem.VibratoOscillator()
	vib.SetWaveform(oscillator.WaveTableSelect(x))
}

func (e SetVibratoWaveform) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
