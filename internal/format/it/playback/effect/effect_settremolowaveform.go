package effect

import (
	"fmt"

	"github.com/gotracker/voice/oscillator"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// SetTremoloWaveform defines a set tremolo waveform effect
type SetTremoloWaveform channel.DataEffect // 'S4x'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetTremoloWaveform) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	x := channel.DataEffect(e) & 0xf

	mem := cs.GetMemory()
	trem := mem.TremoloOscillator()
	trem.SetWaveform(oscillator.WaveTableSelect(x))
	return nil
}

func (e SetTremoloWaveform) String() string {
	return fmt.Sprintf("S%0.2x", channel.DataEffect(e))
}
