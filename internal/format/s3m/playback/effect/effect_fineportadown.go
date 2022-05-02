package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// FinePortaDown defines an fine portamento down effect
type FinePortaDown uint8 // 'EFx'

// Start triggers on the first tick, but before the Tick() function is called
func (e FinePortaDown) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()

	y := uint8(e) & 0x0F

	return doPortaDown(cs, float32(y), 4)
}

func (e FinePortaDown) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
