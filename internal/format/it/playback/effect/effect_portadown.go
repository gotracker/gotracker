package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// PortaDown defines a portamento down effect
type PortaDown uint8 // '2'

// Start triggers on the first tick, but before the Tick() function is called
func (e PortaDown) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
}

// Tick is called on every tick
func (e PortaDown) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	xx := mem.PortaDown(uint8(e))

	if currentTick == 0 {
		return
	}

	doPortaDown(cs, float32(xx), 4, mem.LinearFreqSlides)
}

func (e PortaDown) String() string {
	return fmt.Sprintf("2%0.2x", uint8(e))
}
