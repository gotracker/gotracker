package effect

import (
	"fmt"

	"github.com/gotracker/gotracker/internal/format/s3m/layout/channel"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// FineVolumeSlideUp defines a fine volume slide up effect
type FineVolumeSlideUp ChannelCommand // 'DxF'

// Start triggers on the first tick, but before the Tick() function is called
func (e FineVolumeSlideUp) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e FineVolumeSlideUp) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	x := channel.DataEffect(e) >> 4

	if x != 0x0F && currentTick == 0 {
		return doVolSlide(cs, float32(x), 1.0)
	}
	return nil
}

func (e FineVolumeSlideUp) String() string {
	return fmt.Sprintf("D%0.2x", channel.DataEffect(e))
}
