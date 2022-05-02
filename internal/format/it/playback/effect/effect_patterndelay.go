package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	effectIntf "gotracker/internal/format/it/playback/effect/intf"
	"gotracker/internal/player/intf"
)

// PatternDelay defines a pattern delay effect
type PatternDelay uint8 // 'SEx'

// PreStart triggers when the effect enters onto the channel state
func (e PatternDelay) PreStart(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	m := p.(effectIntf.IT)
	return m.SetPatternDelay(int(uint8(e) & 0x0F))
}

// Start triggers on the first tick, but before the Tick() function is called
func (e PatternDelay) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

func (e PatternDelay) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
