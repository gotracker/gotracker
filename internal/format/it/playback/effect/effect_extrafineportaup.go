package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// ExtraFinePortaUp defines an extra-fine portamento up effect
type ExtraFinePortaUp uint8 // 'FEx'

// Start triggers on the first tick, but before the Tick() function is called
func (e ExtraFinePortaUp) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()

	mem := cs.GetMemory().(*channel.Memory)
	y := mem.PortaUp(uint8(e)) & 0x0F

	return doPortaUp(cs, float32(y), 1, mem.LinearFreqSlides)
}

func (e ExtraFinePortaUp) String() string {
	return fmt.Sprintf("F%0.2x", uint8(e))
}
