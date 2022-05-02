package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	effectIntf "gotracker/internal/format/s3m/playback/effect/intf"
	"gotracker/internal/player/intf"
)

// FinePatternDelay defines an fine pattern delay effect
type FinePatternDelay uint8 // 'S6x'

// Start triggers on the first tick, but before the Tick() function is called
func (e FinePatternDelay) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	m := p.(effectIntf.S3M)
	return m.AddRowTicks(int(x))
}

func (e FinePatternDelay) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
