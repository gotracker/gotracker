package effect

import (
	"fmt"

	"github.com/gotracker/gotracker/internal/format/s3m/layout/channel"
	effectIntf "github.com/gotracker/gotracker/internal/format/s3m/playback/effect/intf"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// FinePatternDelay defines an fine pattern delay effect
type FinePatternDelay ChannelCommand // 'S6x'

// Start triggers on the first tick, but before the Tick() function is called
func (e FinePatternDelay) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	x := channel.DataEffect(e) & 0xf

	m := p.(effectIntf.S3M)
	return m.AddRowTicks(int(x))
}

func (e FinePatternDelay) String() string {
	return fmt.Sprintf("S%0.2x", channel.DataEffect(e))
}
