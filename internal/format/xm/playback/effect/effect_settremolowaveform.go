package effect

import (
	"fmt"

	formatutil "gotracker/internal/format/internal/util"
	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// SetTremoloWaveform defines a set tremolo waveform effect
type SetTremoloWaveform uint8 // 'E7x'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetTremoloWaveform) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	mem := cs.GetMemory().(*channel.Memory)
	trem := mem.TremoloOscillator()
	trem.Table = formatutil.WaveTableSelect(x)
}

func (e SetTremoloWaveform) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
