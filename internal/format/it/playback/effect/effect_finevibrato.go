package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// FineVibrato defines an fine vibrato effect
type FineVibrato uint8 // 'U'

// Start triggers on the first tick, but before the Tick() function is called
func (e FineVibrato) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
}

// Tick is called on every tick
func (e FineVibrato) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	x, y := mem.Vibrato(uint8(e))
	if currentTick != 0 {
		doVibrato(cs, currentTick, x, y, 1)
	}
}

func (e FineVibrato) String() string {
	return fmt.Sprintf("U%0.2x", uint8(e))
}
