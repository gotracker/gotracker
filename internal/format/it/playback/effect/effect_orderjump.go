package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
	"gotracker/internal/song/index"
)

// OrderJump defines an order jump effect
type OrderJump uint8 // 'B'

// Start triggers on the first tick, but before the Tick() function is called
func (e OrderJump) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e OrderJump) Stop(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, lastTick int) error {
	return p.SetNextOrder(index.Order(e))
}

func (e OrderJump) String() string {
	return fmt.Sprintf("B%0.2x", uint8(e))
}
