package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
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
	_, y := mem.VolumeSlide(uint8(e))

	return doVolSlide(cs, -float32(y), 1.0)
}

func (e VolumeSlideDown) String() string {
	return fmt.Sprintf("D%0.2x", uint8(e))
}

//====================================================

// VolChanVolumeSlideDown defines a volume slide down effect (from the volume channel)
type VolChanVolumeSlideDown uint8 // 'd'

// Tick is called on every tick
func (e VolChanVolumeSlideDown) Tick(cs intf.Channel, p intf.Playback, currentTick int) error {
	mem := cs.GetMemory().(*channel.Memory)
	y := mem.VolChanVolumeSlide(uint8(e))

	return doVolSlide(cs, -float32(y), 1.0)
}

func (e VolChanVolumeSlideDown) String() string {
	return fmt.Sprintf("d0%x", uint8(e))
}
