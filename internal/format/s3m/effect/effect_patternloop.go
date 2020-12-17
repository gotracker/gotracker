package effect

import (
	"fmt"

	"gotracker/internal/module/player/intf"
)

// PatternLoop defines a pattern loop effect
type PatternLoop uint8 // 'SBx'

// PreStart triggers when the effect enters onto the channel state
func (e PatternLoop) PreStart(cs intf.Channel, ss intf.Song) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e PatternLoop) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xF

	if x == 0 {
		// set loop
		ss.SetPatternLoopStart()
	} else {
		ss.SetPatternLoopEnd(x)
	}
}

// Tick is called on every tick
func (e PatternLoop) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e PatternLoop) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e PatternLoop) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
