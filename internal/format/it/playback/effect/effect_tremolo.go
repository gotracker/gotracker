package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// Tremolo defines a tremolo effect
type Tremolo uint8 // 'R'

// Start triggers on the first tick, but before the Tick() function is called
func (e Tremolo) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e Tremolo) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	mem := cs.GetMemory()
	x, y := mem.Tremolo(uint8(e))
	// NOTE: JBC - IT dos not update on tick 0, but MOD does.
	// Maybe need to add a flag for converted MOD backward compatibility?
	if currentTick != 0 {
		return doTremolo(cs, currentTick, x, y, 4)
	}
	return nil
}

func (e Tremolo) String() string {
	return fmt.Sprintf("R%0.2x", uint8(e))
}
