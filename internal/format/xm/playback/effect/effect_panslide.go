package effect

import (
	"fmt"

	xmPanning "github.com/gotracker/gotracker/internal/format/xm/conversion/panning"
	"github.com/gotracker/gotracker/internal/format/xm/layout/channel"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// PanSlide defines a pan slide effect
type PanSlide channel.DataEffect // 'Pxx'

// Start triggers on the first tick, but before the Tick() function is called
func (e PanSlide) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	xx := channel.DataEffect(e)
	x := xx >> 4
	y := xx & 0x0F

	xp := channel.DataEffect(xmPanning.PanningToXm(cs.GetPan()))
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
	cs.SetPan(xmPanning.PanningFromXm(uint8(xp)))
	return nil
}

func (e PanSlide) String() string {
	return fmt.Sprintf("P%0.2x", channel.DataEffect(e))
}
