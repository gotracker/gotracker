package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// RowJump defines a row jump effect
type RowJump uint8 // 'C'

// PreStart triggers when the effect enters onto the channel state
func (e RowJump) PreStart(cs intf.Channel, ss intf.Song) {
	ss.SetCurrentRow(uint8((e>>4)*10 + (e & 0x0f)))
}

// Start triggers on the first tick, but before the Tick() function is called
func (e RowJump) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e RowJump) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e RowJump) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e RowJump) String() string {
	return fmt.Sprintf("C%0.2x", uint8(e))
}
