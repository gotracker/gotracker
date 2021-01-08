package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// PatternLoop defines a pattern loop effect
type PatternLoop uint8 // 'E6x'

// PreStart triggers when the effect enters onto the channel state
func (e PatternLoop) PreStart(cs intf.Channel, p intf.Playback) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e PatternLoop) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xF

	mem := cs.GetMemory().(*channel.Memory)
	pl := mem.GetPatternLoop()
	if x == 0 {
		// set loop
		pl.Start = p.GetCurrentRow()
	} else {
		if !pl.Enabled {
			pl.Enabled = true
			pl.Total = x
			pl.End = p.GetCurrentRow()
			pl.Count = 0
		}
		if row, ok := pl.ContinueLoop(p.GetCurrentRow()); ok {
			p.SetNextRow(row)
		}
	}
}

// Tick is called on every tick
func (e PatternLoop) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e PatternLoop) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e PatternLoop) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
