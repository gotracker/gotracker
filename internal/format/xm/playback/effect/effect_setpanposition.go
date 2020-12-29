package effect

import (
	"fmt"

	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/player/intf"
)

// SetFinePanPosition defines a set pan position effect
type SetFinePanPosition uint8 // 'E8x'

// PreStart triggers when the effect enters onto the channel state
func (e SetFinePanPosition) PreStart(cs intf.Channel, p intf.Playback) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetFinePanPosition) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	xy := uint8(e)
	y := xy & 0x0F

	cs.SetPan(util.PanningFromXm(y << 4))
}

// Tick is called on every tick
func (e SetFinePanPosition) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e SetFinePanPosition) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e SetFinePanPosition) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
