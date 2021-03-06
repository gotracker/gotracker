package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
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
	x, _ := mem.VolumeSlide(uint8(e))

	return doVolSlide(cs, float32(x), 1.0)
}

func (e VolumeSlideUp) String() string {
	return fmt.Sprintf("D%0.2x", uint8(e))
}

//====================================================

// VolChanVolumeSlideUp defines a volume slide up effect (from the volume channel)
type VolChanVolumeSlideUp uint8 // 'd'

// Tick is called on every tick
func (e VolChanVolumeSlideUp) Tick(cs intf.Channel, p intf.Playback, currentTick int) error {
	mem := cs.GetMemory().(*channel.Memory)
	x := mem.VolChanVolumeSlide(uint8(e))

	return doVolSlide(cs, float32(x), 1.0)
}

func (e VolChanVolumeSlideUp) String() string {
	return fmt.Sprintf("d%x0", uint8(e))
}
