package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// FineVolumeSlideUp defines a fine volume slide up effect
type FineVolumeSlideUp uint8 // 'D'

// Start triggers on the first tick, but before the Tick() function is called
func (e FineVolumeSlideUp) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e FineVolumeSlideUp) Tick(cs intf.Channel, p intf.Playback, currentTick int) error {
	x := uint8(e) >> 4

	if x != 0x0F && currentTick == 0 {
		return doVolSlide(cs, float32(x), 1.0)
	}
	return nil
}

func (e FineVolumeSlideUp) String() string {
	return fmt.Sprintf("D%0.2x", uint8(e))
}
