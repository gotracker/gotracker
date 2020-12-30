package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// FineVolumeSlideDown defines a volume slide effect
type FineVolumeSlideDown uint8 // 'EAx'

// PreStart triggers when the effect enters onto the channel state
func (e FineVolumeSlideDown) PreStart(cs intf.Channel, p intf.Playback) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e FineVolumeSlideDown) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	mem := cs.GetMemory().(*channel.Memory)
	xy := mem.FineVolumeSlideDown(uint8(e))
	y := uint8(xy & 0x0F)

	doVolSlide(cs, float32(y), 1.0)
}

// Tick is called on every tick
func (e FineVolumeSlideDown) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e FineVolumeSlideDown) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e FineVolumeSlideDown) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
