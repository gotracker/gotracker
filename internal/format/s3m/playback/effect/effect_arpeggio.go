package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// Arpeggio defines an arpeggio effect
type Arpeggio uint8 // 'J'

// PreStart triggers when the effect enters onto the channel state
func (e Arpeggio) PreStart(cs intf.Channel, ss intf.Song) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e Arpeggio) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
	cs.SetPos(cs.GetTargetPos())
}

// Tick is called on every tick
func (e Arpeggio) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	xy := mem.LastNonZero(uint8(e))
	x := int8(xy>>4) - 8
	y := int8(xy&0x0f) - 8
	doArpeggio(cs, currentTick, x, y)
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e Arpeggio) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e Arpeggio) String() string {
	return fmt.Sprintf("J%0.2x", uint8(e))
}
