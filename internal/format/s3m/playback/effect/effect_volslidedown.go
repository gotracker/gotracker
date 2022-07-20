package effect

import (
	"fmt"

	"github.com/gotracker/gotracker/internal/format/s3m/layout/channel"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// VolumeSlideDown defines a volume slide down effect
type VolumeSlideDown ChannelCommand // 'D0y'

// Start triggers on the first tick, but before the Tick() function is called
func (e VolumeSlideDown) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e VolumeSlideDown) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	mem := cs.GetMemory()
	y := channel.DataEffect(e) & 0x0F

	if mem.Shared.VolSlideEveryFrame || currentTick != 0 {
		return doVolSlide(cs, -float32(y), 1.0)
	}
	return nil
}

func (e VolumeSlideDown) String() string {
	return fmt.Sprintf("D%0.2x", channel.DataEffect(e))
}
