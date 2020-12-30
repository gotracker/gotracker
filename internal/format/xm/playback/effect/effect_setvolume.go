package effect

import (
	"fmt"

	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/player/intf"
)

// SetVolume defines a volume slide effect
type SetVolume uint8 // 'C'

// PreStart triggers when the effect enters onto the channel state
func (e SetVolume) PreStart(cs intf.Channel, p intf.Playback) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetVolume) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	xx := uint8(e)

	cs.SetActiveVolume(util.VolumeFromXm(0x10 + xx))
}

// Tick is called on every tick
func (e SetVolume) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e SetVolume) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e SetVolume) String() string {
	return fmt.Sprintf("C%0.2x", uint8(e))
}
