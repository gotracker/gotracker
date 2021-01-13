package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// Tremor defines a tremor effect
type Tremor uint8 // 'I'

// Start triggers on the first tick, but before the Tick() function is called
func (e Tremor) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e Tremor) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	xy := mem.LastNonZero(uint8(e))
	x := int((xy >> 4) + 1)
	y := int((xy & 0x0f) + 1)
	doTremor(cs, currentTick, x, y)
}

func (e Tremor) String() string {
	return fmt.Sprintf("I%0.2x", uint8(e))
}
