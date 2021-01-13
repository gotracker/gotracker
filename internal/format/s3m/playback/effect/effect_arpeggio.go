package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
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
	xy := mem.LastNonZero(uint8(e))
	x := int8(xy >> 4)
	y := int8(xy & 0x0f)
	doArpeggio(cs, currentTick, x, y)
}

func (e Arpeggio) String() string {
	return fmt.Sprintf("J%0.2x", uint8(e))
}
