package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// GlobalVolumeSlide defines a global volume slide effect
type GlobalVolumeSlide uint8 // 'W'

// Start triggers on the first tick, but before the Tick() function is called
func (e GlobalVolumeSlide) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e GlobalVolumeSlide) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	x, y := mem.GlobalVolumeSlide(uint8(e))

	if currentTick == 0 {
		return
	}

	if x == 0 {
		// global vol slide down
		doGlobalVolSlide(p, -float32(y), 1.0)
	} else if y == 0 {
		// global vol slide up
		doGlobalVolSlide(p, float32(y), 1.0)
	}
}

func (e GlobalVolumeSlide) String() string {
	return fmt.Sprintf("W%0.2x", uint8(e))
}
