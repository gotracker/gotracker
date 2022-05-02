package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/player/intf"
)

// SetPanPosition defines a set pan position effect
type SetPanPosition uint8 // 'S8x'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetPanPosition) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	cs.SetPan(util.PanningFromS3M(x))
	return nil
}

func (e SetPanPosition) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
