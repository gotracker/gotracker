package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// ExtraFinePortaUp defines an extra-fine portamento up effect
type ExtraFinePortaUp uint8 // 'X1x'

// Start triggers on the first tick, but before the Tick() function is called
func (e ExtraFinePortaUp) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()

	mem := cs.GetMemory().(*channel.Memory)
	xx := mem.ExtraFinePortaUp(uint8(e))
	y := xx & 0x0F

	return doPortaUp(cs, float32(y), 1, mem.LinearFreqSlides)
}

func (e ExtraFinePortaUp) String() string {
	return fmt.Sprintf("F%0.2x", uint8(e))
}
