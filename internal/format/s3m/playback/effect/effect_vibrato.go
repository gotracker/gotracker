package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// Vibrato defines a vibrato effect
type Vibrato uint8 // 'H'

// Start triggers on the first tick, but before the Tick() function is called
func (e Vibrato) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
	return nil
}

// Tick is called on every tick
func (e Vibrato) Tick(cs intf.Channel, p intf.Playback, currentTick int) error {
	mem := cs.GetMemory().(*channel.Memory)
	x, y := mem.Vibrato(uint8(e))
	// NOTE: JBC - S3M dos not update on tick 0, but MOD does.
	// Maybe need to add a flag for converted MOD backward compatibility?
	if currentTick != 0 {
		return doVibrato(cs, currentTick, x, y, 4)
	}
	return nil
}

func (e Vibrato) String() string {
	return fmt.Sprintf("H%0.2x", uint8(e))
}
