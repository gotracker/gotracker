package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// PortaUp defines a portamento up effect
type PortaUp ChannelCommand // 'F'

// Start triggers on the first tick, but before the Tick() function is called
func (e PortaUp) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
	return nil
}

// Tick is called on every tick
func (e PortaUp) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	mem := cs.GetMemory()
	xx := mem.LastNonZero(channel.DataEffect(e))

	if currentTick != 0 {
		return doPortaUp(cs, float32(xx), 4)
	}
	return nil
}

func (e PortaUp) String() string {
	return fmt.Sprintf("F%0.2x", channel.DataEffect(e))
}
