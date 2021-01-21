package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// Tremolo defines a tremolo effect
type Tremolo uint8 // '7'

// Start triggers on the first tick, but before the Tick() function is called
func (e Tremolo) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e Tremolo) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	xy := mem.Tremolo(uint8(e))
	// NOTE: JBC - IT updates on tick 0, but MOD does not.
	// Just have to eat this incompatibility, I guess...
	x := xy >> 4
	y := xy & 0x0f
	doTremolo(cs, currentTick, x, y, 4)
}

func (e Tremolo) String() string {
	return fmt.Sprintf("7%0.2x", uint8(e))
}
