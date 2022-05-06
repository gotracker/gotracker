package effect

import (
	"fmt"

	"github.com/gotracker/gotracker/internal/format/xm/layout/channel"
	"github.com/gotracker/gotracker/internal/format/xm/playback/util"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// SetVolume defines a volume slide effect
type SetVolume channel.DataEffect // 'C'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetVolume) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	xx := util.VolumeXM(e)

	cs.SetActiveVolume(xx.Volume())
	return nil
}

func (e SetVolume) String() string {
	return fmt.Sprintf("C%0.2x", channel.DataEffect(e))
}
