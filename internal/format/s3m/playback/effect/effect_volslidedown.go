package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// VolumeSlideDown defines a volume slide down effect
type VolumeSlideDown uint8 // 'D'

// Start triggers on the first tick, but before the Tick() function is called
func (e VolumeSlideDown) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e VolumeSlideDown) Tick(cs intf.Channel, p intf.Playback, currentTick int) error {
	mem := cs.GetMemory().(*channel.Memory)
	y := uint8(e) & 0x0F

	if mem.VolSlideEveryFrame || currentTick != 0 {
		return doVolSlide(cs, -float32(y), 1.0)
	}
	return nil
}

func (e VolumeSlideDown) String() string {
	return fmt.Sprintf("D%0.2x", uint8(e))
}
