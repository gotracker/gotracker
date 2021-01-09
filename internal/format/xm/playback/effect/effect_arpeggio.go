package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// Arpeggio defines an arpeggio effect
type Arpeggio uint8 // '0'

// PreStart triggers when the effect enters onto the channel state
func (e Arpeggio) PreStart(cs intf.Channel, p intf.Playback) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e Arpeggio) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
	cs.SetPos(cs.GetTargetPos())
}

// Tick is called on every tick
func (e Arpeggio) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	xy := uint8(e)
	if xy == 0 {
		return
	}

	x := int8(xy >> 4)
	y := int8(xy & 0x0f)
	doArpeggio(cs, currentTick, x, y)
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e Arpeggio) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e Arpeggio) String() string {
	return fmt.Sprintf("0%0.2x", uint8(e))
}
