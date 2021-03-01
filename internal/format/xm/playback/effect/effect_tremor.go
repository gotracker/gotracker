package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// Tremor defines a tremor effect
type Tremor uint8 // 'T'

// Start triggers on the first tick, but before the Tick() function is called
func (e Tremor) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e Tremor) Tick(cs intf.Channel, p intf.Playback, currentTick int) error {
	if currentTick != 0 {
		mem := cs.GetMemory().(*channel.Memory)
		x, y := mem.Tremor(uint8(e))
		return doTremor(cs, currentTick, int(x)+1, int(y)+1)
	}
	return nil
}

func (e Tremor) String() string {
	return fmt.Sprintf("T%0.2x", uint8(e))
}
