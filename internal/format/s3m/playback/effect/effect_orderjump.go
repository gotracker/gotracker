package effect

import (
	"fmt"

	"github.com/gotracker/gotracker/internal/format/s3m/layout/channel"
	"github.com/gotracker/gotracker/internal/player/intf"
	"github.com/gotracker/gotracker/internal/song/index"
)

// OrderJump defines an order jump effect
type OrderJump ChannelCommand // 'B'

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
	return fmt.Sprintf("B%0.2x", channel.DataEffect(e))
}
