package effect

import (
	"fmt"

	"github.com/gotracker/gotracker/internal/format/xm/layout/channel"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// Arpeggio defines an arpeggio effect
type Arpeggio channel.DataEffect // '0'

// Start triggers on the first tick, but before the Tick() function is called
func (e Arpeggio) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
	cs.SetPos(cs.GetTargetPos())
	return nil
}

// Tick is called on every tick
func (e Arpeggio) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	xy := channel.DataEffect(e)
	if xy == 0 {
		return nil
	}

	x := int8(xy >> 4)
	y := int8(xy & 0x0f)
	return doArpeggio(cs, currentTick, x, y)
}

func (e Arpeggio) String() string {
	return fmt.Sprintf("0%0.2x", channel.DataEffect(e))
}
