package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// Arpeggio defines an arpeggio effect
type Arpeggio uint8 // 'J'

// Start triggers on the first tick, but before the Tick() function is called
func (e Arpeggio) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
	cs.SetPos(cs.GetTargetPos())
}

// Tick is called on every tick
func (e Arpeggio) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	x, y := mem.Arpeggio(uint8(e))
	doArpeggio(cs, currentTick, int8(x), int8(y))
}

func (e Arpeggio) String() string {
	return fmt.Sprintf("J%0.2x", uint8(e))
}
