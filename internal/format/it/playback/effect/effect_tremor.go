package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// Tremor defines a tremor effect
type Tremor uint8 // 'I'

// Start triggers on the first tick, but before the Tick() function is called
func (e Tremor) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e Tremor) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	mem := cs.GetMemory()
	x, y := mem.Tremor(uint8(e))
	return doTremor(cs, currentTick, int(x)+1, int(y)+1)
}

func (e Tremor) String() string {
	return fmt.Sprintf("I%0.2x", uint8(e))
}
