package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// FineVolumeSlideUp defines a volume slide effect
type FineVolumeSlideUp uint8 // 'EAx'

// Start triggers on the first tick, but before the Tick() function is called
func (e FineVolumeSlideUp) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	mem := cs.GetMemory()
	xy := mem.FineVolumeSlideUp(uint8(e))
	y := uint8(xy & 0x0F)

	return doVolSlide(cs, float32(y), 1.0)
}

func (e FineVolumeSlideUp) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
