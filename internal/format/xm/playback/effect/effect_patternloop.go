package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// PatternLoop defines a pattern loop effect
type PatternLoop uint8 // 'E6x'

// Start triggers on the first tick, but before the Tick() function is called
func (e PatternLoop) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xF

	mem := cs.GetMemory()
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
			return p.SetNextRow(row, true)
		}
	}
	return nil
}

func (e PatternLoop) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
