package effect

import (
	"fmt"

	"github.com/gotracker/gomixing/sampling"

	"gotracker/internal/player/intf"
)

// RetrigVolumeSlide defines a retriggering volume slide effect
type RetrigVolumeSlide uint8 // 'R'

// Start triggers on the first tick, but before the Tick() function is called
func (e RetrigVolumeSlide) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e RetrigVolumeSlide) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	x := uint8(e) >> 4
	y := uint8(e) & 0x0F
	if y == 0 {
		return
	}

	rt := cs.GetRetriggerCount() + 1
	cs.SetRetriggerCount(rt)
	if rt >= x {
		cs.SetPos(sampling.Pos{})
		cs.ResetRetriggerCount()
		switch x {
		case 1:
			doVolSlide(cs, -1, 1)
		case 2:
			doVolSlide(cs, -2, 1)
		case 3:
			doVolSlide(cs, -4, 1)
		case 4:
			doVolSlide(cs, -8, 1)
		case 5:
			doVolSlide(cs, -6, 1)
		case 6:
			doVolSlideTwoThirds(cs)
		case 7:
			doVolSlide(cs, 0, float32(0.5))
		case 8: // ?
		case 9:
			doVolSlide(cs, 1, 1)
		case 10:
			doVolSlide(cs, 2, 1)
		case 11:
			doVolSlide(cs, 4, 1)
		case 12:
			doVolSlide(cs, 8, 1)
		case 13:
			doVolSlide(cs, 16, 1)
		case 14:
			doVolSlide(cs, 0, float32(1.5))
		case 15:
			doVolSlide(cs, 0, 2)
		}
	}
}

func (e RetrigVolumeSlide) String() string {
	return fmt.Sprintf("R%0.2x", uint8(e))
}
