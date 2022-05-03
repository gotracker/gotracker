package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// Tremor defines a tremor effect
type Tremor channel.DataEffect // 'T'

// Start triggers on the first tick, but before the Tick() function is called
func (e Tremor) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e Tremor) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	if currentTick != 0 {
		mem := cs.GetMemory()
		x, y := mem.Tremor(channel.DataEffect(e))
		return doTremor(cs, currentTick, int(x)+1, int(y)+1)
	}
	return nil
}

func (e Tremor) String() string {
	return fmt.Sprintf("T%0.2x", channel.DataEffect(e))
}
