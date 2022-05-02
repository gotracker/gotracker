package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// PortaDown defines a portamento down effect
type PortaDown uint8 // 'E'

// Start triggers on the first tick, but before the Tick() function is called
func (e PortaDown) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
	return nil
}

// Tick is called on every tick
func (e PortaDown) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	mem := cs.GetMemory()
	xx := mem.LastNonZero(uint8(e))

	if currentTick != 0 {
		return doPortaDown(cs, float32(xx), 4)
	}
	return nil
}

func (e PortaDown) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
