package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// PatternDelay defines a pattern delay effect
type PatternDelay uint8 // 'SEx'

// PreStart triggers when the effect enters onto the channel state
func (e PatternDelay) PreStart(cs intf.Channel, ss intf.Song) {
	ss.SetPatternDelay(int(uint8(e) & 0x0F))
}

// Start triggers on the first tick, but before the Tick() function is called
func (e PatternDelay) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e PatternDelay) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e PatternDelay) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e PatternDelay) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
