package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// PortaUp defines a portamento up effect
type PortaUp uint8 // '1'

// Start triggers on the first tick, but before the Tick() function is called
func (e PortaUp) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
	return nil
}

// Tick is called on every tick
func (e PortaUp) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	mem := cs.GetMemory()
	xx := mem.PortaUp(uint8(e))

	if currentTick == 0 {
		return nil
	}

	return doPortaUp(cs, float32(xx), 4, mem.LinearFreqSlides)
}

func (e PortaUp) String() string {
	return fmt.Sprintf("1%0.2x", uint8(e))
}
