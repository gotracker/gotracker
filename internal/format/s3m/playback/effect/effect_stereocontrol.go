package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/player/intf"
)

// StereoControl defines a set stereo control effect
type StereoControl ChannelCommand // 'SAx'

// Start triggers on the first tick, but before the Tick() function is called
func (e StereoControl) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	if x > 7 {
		cs.SetPan(util.PanningFromS3M(x - 8))
	} else {
		cs.SetPan(util.PanningFromS3M(x + 8))
	}
	return nil
}

func (e StereoControl) String() string {
	return fmt.Sprintf("S%0.2x", channel.DataEffect(e))
}
