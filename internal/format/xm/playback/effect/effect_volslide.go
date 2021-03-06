package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// VolumeSlide defines a volume slide effect
type VolumeSlide uint8 // 'A'

// Start triggers on the first tick, but before the Tick() function is called
func (e VolumeSlide) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e VolumeSlide) Tick(cs intf.Channel, p intf.Playback, currentTick int) error {
	mem := cs.GetMemory().(*channel.Memory)
	x, y := mem.VolumeSlide(uint8(e))

	if currentTick == 0 {
		return nil
	}

	if x == 0 {
		// vol slide down
		return doVolSlide(cs, -float32(y), 1.0)
	} else if y == 0 {
		// vol slide up
		return doVolSlide(cs, float32(y), 1.0)
	}
	return nil
}

func (e VolumeSlide) String() string {
	return fmt.Sprintf("A%0.2x", uint8(e))
}
