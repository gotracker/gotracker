package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// FineVolumeSlideUp defines a fine volume slide up effect
type FineVolumeSlideUp channel.DataEffect // 'D'

// Start triggers on the first tick, but before the Tick() function is called
func (e FineVolumeSlideUp) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e FineVolumeSlideUp) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	mem := cs.GetMemory()
	x, _ := mem.VolumeSlide(channel.DataEffect(e))

	if x != 0x0F && currentTick == 0 {
		return doVolSlide(cs, float32(x), 1.0)
	}
	return nil
}

func (e FineVolumeSlideUp) String() string {
	return fmt.Sprintf("D%0.2x", channel.DataEffect(e))
}

//====================================================

// VolChanFineVolumeSlideUp defines a fine volume slide up effect (from the volume channel)
type VolChanFineVolumeSlideUp channel.DataEffect // 'd'

// Start triggers on the first tick, but before the Tick() function is called
func (e VolChanFineVolumeSlideUp) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	mem := cs.GetMemory()
	x := mem.VolChanVolumeSlide(channel.DataEffect(e))

	return doVolSlide(cs, float32(x), 1.0)
}

func (e VolChanFineVolumeSlideUp) String() string {
	return fmt.Sprintf("d%xF", channel.DataEffect(e))
}
