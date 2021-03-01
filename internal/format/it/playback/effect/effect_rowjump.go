package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// RowJump defines a row jump effect
type RowJump uint8 // 'C'

// Start triggers on the first tick, but before the Tick() function is called
func (e RowJump) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e RowJump) Stop(cs intf.Channel, p intf.Playback, lastTick int) error {
	r := uint8(e)
	rowIdx := intf.RowIdx(r)
	return p.SetNextRow(rowIdx)
}

func (e RowJump) String() string {
	return fmt.Sprintf("C%0.2x", uint8(e))
}
