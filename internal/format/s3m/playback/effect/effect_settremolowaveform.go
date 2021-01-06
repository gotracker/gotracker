package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// SetTremoloWaveform defines a set tremolo waveform effect
type SetTremoloWaveform uint8 // 'S4x'

// PreStart triggers when the effect enters onto the channel state
func (e SetTremoloWaveform) PreStart(cs intf.Channel, p intf.Playback) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetTremoloWaveform) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	mem := cs.GetMemory().(*channel.Memory)
	trem := mem.TremoloOscillator()
	trem.Table = channel.WaveTableSelect(x)
}

// Tick is called on every tick
func (e SetTremoloWaveform) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e SetTremoloWaveform) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e SetTremoloWaveform) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
