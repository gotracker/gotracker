package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// FinePortaUp defines an fine portamento up effect
type FinePortaUp ChannelCommand // 'FFx'

// Start triggers on the first tick, but before the Tick() function is called
func (e FinePortaUp) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()

	y := channel.DataEffect(e) & 0x0F

	return doPortaUp(cs, float32(y), 4)
}

func (e FinePortaUp) String() string {
	return fmt.Sprintf("F%0.2x", channel.DataEffect(e))
}
