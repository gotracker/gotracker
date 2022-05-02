package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// FineVolumeSlideDown defines a fine volume slide down effect
type FineVolumeSlideDown uint8 // 'D'

// Start triggers on the first tick, but before the Tick() function is called
func (e FineVolumeSlideDown) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e FineVolumeSlideDown) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	mem := cs.GetMemory()
	_, y := mem.VolumeSlide(uint8(e))

	if y != 0x0F && currentTick == 0 {
		return doVolSlide(cs, -float32(y), 1.0)
	}
	return nil
}

func (e FineVolumeSlideDown) String() string {
	return fmt.Sprintf("D%0.2x", uint8(e))
}

//====================================================

// VolChanFineVolumeSlideDown defines a fine volume slide down effect (from the volume channel)
type VolChanFineVolumeSlideDown uint8 // 'd'

// Start triggers on the first tick, but before the Tick() function is called
func (e VolChanFineVolumeSlideDown) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	mem := cs.GetMemory()
	y := mem.VolChanVolumeSlide(uint8(e))

	return doVolSlide(cs, -float32(y), 1.0)
}

func (e VolChanFineVolumeSlideDown) String() string {
	return fmt.Sprintf("dF%x", uint8(e))
}
