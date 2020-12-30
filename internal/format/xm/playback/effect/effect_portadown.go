package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// PortaDown defines a portamento down effect
type PortaDown uint8 // '2'

// PreStart triggers when the effect enters onto the channel state
func (e PortaDown) PreStart(cs intf.Channel, p intf.Playback) {
	cs.SetKeepFinetune(true)
}

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

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e PortaDown) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e PortaDown) String() string {
	return fmt.Sprintf("2%0.2x", uint8(e))
}
