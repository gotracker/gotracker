package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// ExtraFinePortaDown defines an extra-fine portamento down effect
type ExtraFinePortaDown uint8 // 'EEx'

// Start triggers on the first tick, but before the Tick() function is called
func (e ExtraFinePortaDown) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()

	mem := cs.GetMemory().(*channel.Memory)
	y := mem.PortaDown(uint8(e)) & 0x0F

	return doPortaDown(cs, float32(y), 1, mem.LinearFreqSlides)
}

func (e ExtraFinePortaDown) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
