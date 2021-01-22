package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// Vibrato defines a vibrato effect
type Vibrato uint8 // 'H'

// Start triggers on the first tick, but before the Tick() function is called
func (e Vibrato) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
}

// Tick is called on every tick
func (e Vibrato) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	x, y := mem.Vibrato(uint8(e))
	if mem.OldEffectMode {
		if currentTick != 0 {
			doVibrato(cs, currentTick, x, y, 8)
		}
	} else {
		doVibrato(cs, currentTick, x, y, 4)
	}
}

func (e Vibrato) String() string {
	return fmt.Sprintf("H%0.2x", uint8(e))
}
