package effect

import (
	"fmt"

	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/player/intf"
)

// SetGlobalVolume defines a set global volume effect
type SetGlobalVolume uint8 // 'G'

// PreStart triggers when the effect enters onto the channel state
func (e SetGlobalVolume) PreStart(cs intf.Channel, p intf.Playback) {
	p.SetGlobalVolume(util.VolumeFromXm(uint8(e)))
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetGlobalVolume) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e SetGlobalVolume) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e SetGlobalVolume) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e SetGlobalVolume) String() string {
	return fmt.Sprintf("G%0.2x", uint8(e))
}
