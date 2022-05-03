package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// VolumeSlideUp defines a volume slide up effect
type VolumeSlideUp channel.DataEffect // 'D'

// Start triggers on the first tick, but before the Tick() function is called
func (e VolumeSlideUp) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e VolumeSlideUp) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	mem := cs.GetMemory()
	x, _ := mem.VolumeSlide(channel.DataEffect(e))

	return doVolSlide(cs, float32(x), 1.0)
}

func (e VolumeSlideUp) String() string {
	return fmt.Sprintf("D%0.2x", channel.DataEffect(e))
}

//====================================================

// VolChanVolumeSlideUp defines a volume slide up effect (from the volume channel)
type VolChanVolumeSlideUp channel.DataEffect // 'd'

// Tick is called on every tick
func (e VolChanVolumeSlideUp) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	mem := cs.GetMemory()
	x := mem.VolChanVolumeSlide(channel.DataEffect(e))

	return doVolSlide(cs, float32(x), 1.0)
}

func (e VolChanVolumeSlideUp) String() string {
	return fmt.Sprintf("d%x0", channel.DataEffect(e))
}
