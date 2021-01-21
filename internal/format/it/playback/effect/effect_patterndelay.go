package effect

import (
	"fmt"

	effectIntf "gotracker/internal/format/it/playback/effect/intf"
	"gotracker/internal/player/intf"
)

// PatternDelay defines a pattern delay effect
type PatternDelay uint8 // 'SEx'

// PreStart triggers when the effect enters onto the channel state
func (e PatternDelay) PreStart(cs intf.Channel, p intf.Playback) {
	m := p.(effectIntf.IT)
	m.SetPatternDelay(int(uint8(e) & 0x0F))
}

// Start triggers on the first tick, but before the Tick() function is called
func (e PatternDelay) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
}

func (e PatternDelay) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
