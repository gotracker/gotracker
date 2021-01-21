package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// PortaUp defines a portamento up effect
type PortaUp uint8 // '1'

// Start triggers on the first tick, but before the Tick() function is called
func (e PortaUp) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
}

// Tick is called on every tick
func (e PortaUp) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	xx := mem.PortaUp(uint8(e))

	if currentTick == 0 {
		return
	}

	doPortaUp(cs, float32(xx), 4, mem.LinearFreqSlides)
}

func (e PortaUp) String() string {
	return fmt.Sprintf("1%0.2x", uint8(e))
}
