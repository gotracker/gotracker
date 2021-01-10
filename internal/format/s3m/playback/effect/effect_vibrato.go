package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// Vibrato defines a vibrato effect
type Vibrato uint8 // 'H'

// PreStart triggers when the effect enters onto the channel state
func (e Vibrato) PreStart(cs intf.Channel, p intf.Playback) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e Vibrato) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
}

// Tick is called on every tick
func (e Vibrato) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	xy := mem.Vibrato(uint8(e))
	// NOTE: JBC - S3M dos not update on tick 0, but MOD does.
	// Maybe need to add a flag for converted MOD backward compatibility?
	if currentTick == 0 {
		x := xy >> 4
		y := xy & 0x0f
		doVibrato(cs, currentTick, x, y, 4)
	}
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e Vibrato) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e Vibrato) String() string {
	return fmt.Sprintf("H%0.2x", uint8(e))
}
