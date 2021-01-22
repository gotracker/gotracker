package effect

import (
	"fmt"

	effectIntf "gotracker/internal/format/it/playback/effect/intf"
	"gotracker/internal/player/intf"
)

// SetSpeed defines a set speed effect
type SetSpeed uint8 // 'A'

// PreStart triggers when the effect enters onto the channel state
func (e SetSpeed) PreStart(cs intf.Channel, p intf.Playback) {
	if e != 0 {
		m := p.(effectIntf.IT)
		m.SetTicks(int(e))
	}
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetSpeed) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
}

func (e SetSpeed) String() string {
	return fmt.Sprintf("A%0.2x", uint8(e))
}
