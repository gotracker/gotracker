package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// Tremolo defines a tremolo effect
type Tremolo uint8 // 'R'

// PreStart triggers when the effect enters onto the channel state
func (e Tremolo) PreStart(cs intf.Channel, p intf.Playback) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e Tremolo) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e Tremolo) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	xy := mem.LastNonZero(uint8(e))
	// NOTE: JBC - S3M updates on tick 0, but MOD does not.
	// Maybe need to add a flag for converted MOD backward compatibility?
	x := xy >> 4
	y := xy & 0x0f
	doTremolo(cs, currentTick, x, y, 4)
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e Tremolo) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e Tremolo) String() string {
	return fmt.Sprintf("R%0.2x", uint8(e))
}
