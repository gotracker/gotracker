package effect

import (
	"fmt"

	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/player/intf"
)

// SetGlobalVolume defines a set global volume effect
type SetGlobalVolume uint8 // 'G'

// PreStart triggers when the effect enters onto the channel state
func (e SetGlobalVolume) PreStart(cs intf.Channel, p intf.Playback) {
	v := util.VolumeXM(e)
	p.SetGlobalVolume(v.Volume())
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetGlobalVolume) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
}

func (e SetGlobalVolume) String() string {
	return fmt.Sprintf("G%0.2x", uint8(e))
}
