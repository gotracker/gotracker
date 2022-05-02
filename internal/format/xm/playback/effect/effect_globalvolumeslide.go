package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// GlobalVolumeSlide defines a global volume slide effect
type GlobalVolumeSlide uint8 // 'H'

// Start triggers on the first tick, but before the Tick() function is called
func (e GlobalVolumeSlide) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e GlobalVolumeSlide) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	mem := cs.GetMemory()
	x, y := mem.GlobalVolumeSlide(uint8(e))

	if currentTick == 0 {
		return nil
	}

	if x == 0 {
		// global vol slide down
		return doGlobalVolSlide(p, -float32(y), 1.0)
	} else if y == 0 {
		// global vol slide up
		return doGlobalVolSlide(p, float32(y), 1.0)
	}
	return nil
}

func (e GlobalVolumeSlide) String() string {
	return fmt.Sprintf("H%0.2x", uint8(e))
}
