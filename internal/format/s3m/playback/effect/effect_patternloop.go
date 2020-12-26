package effect

import (
	"fmt"

	effectIntf "gotracker/internal/format/s3m/playback/effect/intf"
	"gotracker/internal/player/intf"
)

// PatternLoop defines a pattern loop effect
type PatternLoop uint8 // 'SBx'

// PreStart triggers when the effect enters onto the channel state
func (e PatternLoop) PreStart(cs intf.Channel, p intf.Playback) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e PatternLoop) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xF

	m := p.(effectIntf.S3M)
	if x == 0 {
		// set loop
		m.SetPatternLoopStart()
	} else {
		m.SetPatternLoopEnd()
		m.SetPatternLoopCount(int(x))
	}
}

// Tick is called on every tick
func (e PatternLoop) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e PatternLoop) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e PatternLoop) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
