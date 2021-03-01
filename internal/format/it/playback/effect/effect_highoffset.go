package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// HighOffset defines a sample high offset effect
type HighOffset uint8 // 'SAx'

// Start triggers on the first tick, but before the Tick() function is called
func (e HighOffset) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()
	mem := cs.GetMemory().(*channel.Memory)

	xx := uint8(e)

	mem.HighOffset = int(xx) * 0x10000
	return nil
}

func (e HighOffset) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
