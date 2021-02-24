package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// VolumeSlide defines a volume slide effect
type VolumeSlide uint8 // 'A'

// Start triggers on the first tick, but before the Tick() function is called
func (e VolumeSlide) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e VolumeSlide) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	x, y := mem.VolumeSlide(uint8(e))

	if currentTick == 0 {
		return
	}

	if x == 0 {
		// vol slide down
		doVolSlide(cs, -float32(y), 1.0)
	} else if y == 0 {
		// vol slide up
		doVolSlide(cs, float32(y), 1.0)
	}
}

func (e VolumeSlide) String() string {
	return fmt.Sprintf("A%0.2x", uint8(e))
}
