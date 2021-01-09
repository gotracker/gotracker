package effect

import (
	"fmt"

	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/player/intf"
)

// PanSlide defines a pan slide effect
type PanSlide uint8 // 'Pxx'

// PreStart triggers when the effect enters onto the channel state
func (e PanSlide) PreStart(cs intf.Channel, p intf.Playback) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e PanSlide) Start(cs intf.Channel, p intf.Playback) {
	xx := uint8(e)
	x := xx >> 4
	y := xx & 0x0F

	xp := util.PanningToXm(cs.GetPan())
	if x == 0 {
		// slide left y units
		if xp < y {
			xp = 0
		} else {
			xp -= y
		}
	} else if y == 0 {
		// slide right x units
		if xp > 0xFF-x {
			xp = 0xFF
		} else {
			xp += x
		}
	}
	cs.SetPan(util.PanningFromXm(xp))
}

// Tick is called on every tick
func (e PanSlide) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e PanSlide) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e PanSlide) String() string {
	return fmt.Sprintf("P%0.2x", uint8(e))
}
