package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// FineVolumeSlideUp defines a volume slide effect
type FineVolumeSlideUp uint8 // 'EAx'

// PreStart triggers when the effect enters onto the channel state
func (e FineVolumeSlideUp) PreStart(cs intf.Channel, p intf.Playback) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e FineVolumeSlideUp) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	mem := cs.GetMemory().(*channel.Memory)
	xy := mem.FineVolumeSlideUp(uint8(e))
	y := uint8(xy & 0x0F)

	doVolSlide(cs, float32(y), 1.0)
}

// Tick is called on every tick
func (e FineVolumeSlideUp) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e FineVolumeSlideUp) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e FineVolumeSlideUp) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}