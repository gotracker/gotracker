package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// VolumeSlideUp defines a volume slide up effect
type VolumeSlideUp uint8 // 'D'

// Start triggers on the first tick, but before the Tick() function is called
func (e VolumeSlideUp) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e VolumeSlideUp) Tick(cs intf.Channel, p intf.Playback, currentTick int) error {
	mem := cs.GetMemory().(*channel.Memory)
	x := uint8(e) >> 4

	if mem.VolSlideEveryFrame || currentTick != 0 {
		return doVolSlide(cs, float32(x), 1.0)
	}
	return nil
}

func (e VolumeSlideUp) String() string {
	return fmt.Sprintf("D%0.2x", uint8(e))
}
