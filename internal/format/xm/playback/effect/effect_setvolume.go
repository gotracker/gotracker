package effect

import (
	"fmt"

	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/player/intf"
)

// SetVolume defines a volume slide effect
type SetVolume uint8 // 'C'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetVolume) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()

	xx := util.VolumeXM(e)

	cs.SetActiveVolume(xx.Volume())
	return nil
}

func (e SetVolume) String() string {
	return fmt.Sprintf("C%0.2x", uint8(e))
}
