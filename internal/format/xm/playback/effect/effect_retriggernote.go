package effect

import (
	"fmt"

	"github.com/gotracker/gomixing/sampling"

	"gotracker/internal/player/intf"
)

// RetriggerNote defines a retriggering effect
type RetriggerNote uint8 // 'E9x'

// Start triggers on the first tick, but before the Tick() function is called
func (e RetriggerNote) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e RetriggerNote) Tick(cs intf.Channel, p intf.Playback, currentTick int) error {
	y := uint8(e) & 0x0F
	if y == 0 {
		return nil
	}

	rt := cs.GetRetriggerCount() + 1
	cs.SetRetriggerCount(rt)
	if rt >= y {
		cs.SetPos(sampling.Pos{})
		cs.ResetRetriggerCount()
	}
	return nil
}

func (e RetriggerNote) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
