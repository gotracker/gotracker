package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
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
	mem := cs.GetMemory().(*channel.Memory)
	x, _ := mem.VolumeSlide(uint8(e))

	if x != 0x0F && currentTick == 0 {
		return doVolSlide(cs, float32(x), 1.0)
	}
	return nil
}

func (e FineVolumeSlideUp) String() string {
	return fmt.Sprintf("D%0.2x", uint8(e))
}

//====================================================

// VolChanFineVolumeSlideUp defines a fine volume slide up effect (from the volume channel)
type VolChanFineVolumeSlideUp uint8 // 'd'

// Start triggers on the first tick, but before the Tick() function is called
func (e VolChanFineVolumeSlideUp) Start(cs intf.Channel, p intf.Playback) error {
	mem := cs.GetMemory().(*channel.Memory)
	x := mem.VolChanVolumeSlide(uint8(e))

	return doVolSlide(cs, float32(x), 1.0)
}

func (e VolChanFineVolumeSlideUp) String() string {
	return fmt.Sprintf("d%xF", uint8(e))
}
