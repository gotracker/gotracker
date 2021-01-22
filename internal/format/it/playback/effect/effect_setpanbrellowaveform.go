package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/oscillator"
	"gotracker/internal/player/intf"
)

// SetPanbrelloWaveform defines a set panbrello waveform effect
type SetPanbrelloWaveform uint8 // 'S5x'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetPanbrelloWaveform) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	mem := cs.GetMemory().(*channel.Memory)
	panb := mem.PanbrelloOscillator()
	panb.Table = oscillator.WaveTableSelect(x)
}

func (e SetPanbrelloWaveform) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
