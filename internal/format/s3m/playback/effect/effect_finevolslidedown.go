package effect

import (
	"fmt"

	"github.com/gotracker/gotracker/internal/format/s3m/layout/channel"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// FineVolumeSlideDown defines a fine volume slide down effect
type FineVolumeSlideDown ChannelCommand // 'DFy'

// Start triggers on the first tick, but before the Tick() function is called
func (e FineVolumeSlideDown) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e FineVolumeSlideDown) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	y := channel.DataEffect(e) & 0x0F

	if y != 0x0F && currentTick == 0 {
		return doVolSlide(cs, -float32(y), 1.0)
	}
	return nil
}

func (e FineVolumeSlideDown) String() string {
	return fmt.Sprintf("D%0.2x", channel.DataEffect(e))
}
