package effect

import (
	"fmt"

	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/player/intf"
)

// SetPanPosition defines a set pan position effect
type SetPanPosition uint8 // '8xx'

// PreStart triggers when the effect enters onto the channel state
func (e SetPanPosition) PreStart(cs intf.Channel, p intf.Playback) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetPanPosition) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	xx := uint8(e)

	cs.SetPan(util.PanningFromXm(xx))
}

// Tick is called on every tick
func (e SetPanPosition) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e SetPanPosition) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e SetPanPosition) String() string {
	return fmt.Sprintf("8%0.2x", uint8(e))
}
