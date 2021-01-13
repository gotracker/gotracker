package effect

import (
	"fmt"

	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/player/intf"
)

// SetVolume defines a volume slide effect
type SetVolume uint8 // 'C'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetVolume) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	xx := uint8(e)

	cs.SetActiveVolume(util.VolumeFromXm(0x10 + xx))
}

func (e SetVolume) String() string {
	return fmt.Sprintf("C%0.2x", uint8(e))
}
