package effect

import (
	"fmt"

	"github.com/gotracker/voice/oscillator"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// SetTremoloWaveform defines a set tremolo waveform effect
type SetTremoloWaveform uint8 // 'S4x'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetTremoloWaveform) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	mem := cs.GetMemory()
	trem := mem.TremoloOscillator()
	trem.SetWaveform(oscillator.WaveTableSelect(x))
	return nil
}

func (e SetTremoloWaveform) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
