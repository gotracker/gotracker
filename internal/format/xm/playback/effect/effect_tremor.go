package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// Tremor defines a tremor effect
type Tremor uint8 // 'T'

// Start triggers on the first tick, but before the Tick() function is called
func (e Tremor) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e Tremor) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	if currentTick != 0 {
		mem := cs.GetMemory().(*channel.Memory)
		x, y := mem.Tremor(uint8(e))
		doTremor(cs, currentTick, int(x)+1, int(y)+1)
	}
}

func (e Tremor) String() string {
	return fmt.Sprintf("T%0.2x", uint8(e))
}
