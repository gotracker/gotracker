package effect

import (
	"fmt"

	"github.com/gotracker/gomixing/sampling"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// RetrigVolumeSlide defines a retriggering volume slide effect
type RetrigVolumeSlide uint8 // 'R'

// Start triggers on the first tick, but before the Tick() function is called
func (e RetrigVolumeSlide) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e RetrigVolumeSlide) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	x := uint8(e) >> 4
	y := uint8(e) & 0x0F
	if y == 0 {
		return nil
	}

	rt := cs.GetRetriggerCount() + 1
	cs.SetRetriggerCount(rt)
	if rt >= x {
		cs.SetPos(sampling.Pos{})
		cs.ResetRetriggerCount()
		switch x {
		case 1:
			return doVolSlide(cs, -1, 1)
		case 2:
			return doVolSlide(cs, -2, 1)
		case 3:
			return doVolSlide(cs, -4, 1)
		case 4:
			return doVolSlide(cs, -8, 1)
		case 5:
			return doVolSlide(cs, -6, 1)
		case 6:
			return doVolSlideTwoThirds(cs)
		case 7:
			return doVolSlide(cs, 0, float32(0.5))
		case 8: // ?
		case 9:
			return doVolSlide(cs, 1, 1)
		case 10:
			return doVolSlide(cs, 2, 1)
		case 11:
			return doVolSlide(cs, 4, 1)
		case 12:
			return doVolSlide(cs, 8, 1)
		case 13:
			return doVolSlide(cs, 16, 1)
		case 14:
			return doVolSlide(cs, 0, float32(1.5))
		case 15:
			return doVolSlide(cs, 0, 2)
		}
	}
	return nil
}

func (e RetrigVolumeSlide) String() string {
	return fmt.Sprintf("R%0.2x", uint8(e))
}
