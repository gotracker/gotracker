package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/player/intf"
)

// PanSlide defines a pan slide effect
type PanSlide uint8 // 'Pxx'

// Start triggers on the first tick, but before the Tick() function is called
func (e PanSlide) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
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
	return nil
}

func (e PanSlide) String() string {
	return fmt.Sprintf("P%0.2x", uint8(e))
}
