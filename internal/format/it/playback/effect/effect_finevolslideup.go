package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// FineVolumeSlideUp defines a fine volume slide up effect
type FineVolumeSlideUp uint8 // 'D'

// Start triggers on the first tick, but before the Tick() function is called
func (e FineVolumeSlideUp) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e FineVolumeSlideUp) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	x, _ := mem.VolumeSlide(uint8(e))

	if x != 0x0F && currentTick == 0 {
		doVolSlide(cs, float32(x), 1.0)
	}
}

func (e FineVolumeSlideUp) String() string {
	return fmt.Sprintf("D%0.2x", uint8(e))
}
