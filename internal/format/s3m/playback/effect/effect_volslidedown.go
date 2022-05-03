package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// VolumeSlideDown defines a volume slide down effect
type VolumeSlideDown ChannelCommand // 'D'

// Start triggers on the first tick, but before the Tick() function is called
func (e VolumeSlideDown) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e VolumeSlideDown) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	mem := cs.GetMemory()
	y := channel.DataEffect(e) & 0x0F

	if mem.VolSlideEveryFrame || currentTick != 0 {
		return doVolSlide(cs, -float32(y), 1.0)
	}
	return nil
}

func (e VolumeSlideDown) String() string {
	return fmt.Sprintf("D%0.2x", channel.DataEffect(e))
}
