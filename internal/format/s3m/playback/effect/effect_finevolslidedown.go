package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// FineVolumeSlideDown defines a fine volume slide down effect
type FineVolumeSlideDown uint8 // 'D'

// Start triggers on the first tick, but before the Tick() function is called
func (e FineVolumeSlideDown) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e FineVolumeSlideDown) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	y := uint8(e) & 0x0F

	if y != 0x0F && currentTick == 0 {
		doVolSlide(cs, -float32(y), 1.0)
	}
}

func (e FineVolumeSlideDown) String() string {
	return fmt.Sprintf("D%0.2x", uint8(e))
}
