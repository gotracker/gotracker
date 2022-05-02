package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	effectIntf "gotracker/internal/format/it/playback/effect/intf"
	"gotracker/internal/player/intf"
)

// FinePatternDelay defines an fine pattern delay effect
type FinePatternDelay uint8 // 'S6x'

// Start triggers on the first tick, but before the Tick() function is called
func (e FinePatternDelay) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	m := p.(effectIntf.IT)
	if err := m.AddRowTicks(int(x)); err != nil {
		return err
	}
	return nil
}

func (e FinePatternDelay) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
