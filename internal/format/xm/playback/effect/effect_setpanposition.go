package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/player/intf"
)

// SetPanPosition defines a set pan position effect
type SetPanPosition channel.DataEffect // '8xx'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetPanPosition) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	xx := uint8(e)

	cs.SetPan(util.PanningFromXm(xx))
	return nil
}

func (e SetPanPosition) String() string {
	return fmt.Sprintf("8%0.2x", channel.DataEffect(e))
}
