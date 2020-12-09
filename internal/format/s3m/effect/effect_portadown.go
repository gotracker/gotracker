package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
)

// PortaDown defines a portamento down effect
type PortaDown uint8 // 'E'

// PreStart triggers when the effect enters onto the channel state
func (e PortaDown) PreStart(cs intf.Channel, ss intf.Song) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e PortaDown) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
}

// Tick is called on every tick
func (e PortaDown) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	xx := cs.GetEffectSharedMemory(uint8(e))

	if currentTick != 0 {
		doPortaDown(cs, float32(xx), 4)
	}
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e PortaDown) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e PortaDown) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
