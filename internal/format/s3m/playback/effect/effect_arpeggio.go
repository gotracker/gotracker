package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// Arpeggio defines an arpeggio effect
type Arpeggio uint8 // 'J'

// Start triggers on the first tick, but before the Tick() function is called
func (e Arpeggio) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
	cs.SetPos(cs.GetTargetPos())
	return nil
}

// Tick is called on every tick
func (e Arpeggio) Tick(cs intf.Channel, p intf.Playback, currentTick int) error {
	mem := cs.GetMemory().(*channel.Memory)
	x, y := mem.LastNonZeroXY(uint8(e))
	return doArpeggio(cs, currentTick, int8(x), int8(y))
}

func (e Arpeggio) String() string {
	return fmt.Sprintf("J%0.2x", uint8(e))
}
