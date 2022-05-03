package effect

import (
	"fmt"

	"github.com/gotracker/voice/oscillator"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// SetPanbrelloWaveform defines a set panbrello waveform effect
type SetPanbrelloWaveform channel.DataEffect // 'S5x'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetPanbrelloWaveform) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	x := channel.DataEffect(e) & 0xf

	mem := cs.GetMemory()
	panb := mem.PanbrelloOscillator()
	panb.SetWaveform(oscillator.WaveTableSelect(x))
	return nil
}

func (e SetPanbrelloWaveform) String() string {
	return fmt.Sprintf("S%0.2x", channel.DataEffect(e))
}
