package effect

import (
	"fmt"

	itfile "github.com/gotracker/goaudiofile/music/tracked/it"

	"gotracker/internal/format/it/playback/util"
	"gotracker/internal/player/intf"
)

// PanSlide defines a pan slide effect
type PanSlide uint8 // 'Pxx'

// Start triggers on the first tick, but before the Tick() function is called
func (e PanSlide) Start(cs intf.Channel, p intf.Playback) {
	xx := uint8(e)
	x := itfile.PanValue(xx >> 4)
	y := itfile.PanValue(xx & 0x0F)

	xp := util.PanningToIt(cs.GetPan())
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
	cs.SetPan(util.PanningFromIt(xp))
}

func (e PanSlide) String() string {
	return fmt.Sprintf("P%0.2x", uint8(e))
}
