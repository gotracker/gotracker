package effect

import (
	"fmt"

	s3mPanning "github.com/gotracker/gotracker/internal/format/s3m/conversion/panning"
	"github.com/gotracker/gotracker/internal/format/s3m/layout/channel"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// SetPanPosition defines a set pan position effect
type SetPanPosition ChannelCommand // 'S8x'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetPanPosition) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	cs.SetPan(s3mPanning.PanningFromS3M(x))
	return nil
}

func (e SetPanPosition) String() string {
	return fmt.Sprintf("S%0.2x", channel.DataEffect(e))
}
