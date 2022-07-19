package effect

import (
	"fmt"

	s3mPanning "github.com/gotracker/gotracker/internal/format/s3m/conversion/panning"
	"github.com/gotracker/gotracker/internal/format/s3m/layout/channel"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// StereoControl defines a set stereo control effect
type StereoControl ChannelCommand // 'SAx'

// Start triggers on the first tick, but before the Tick() function is called
func (e StereoControl) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	if x > 7 {
		cs.SetPan(s3mPanning.PanningFromS3M(x - 8))
	} else {
		cs.SetPan(s3mPanning.PanningFromS3M(x + 8))
	}
	return nil
}

func (e StereoControl) String() string {
	return fmt.Sprintf("S%0.2x", channel.DataEffect(e))
}
