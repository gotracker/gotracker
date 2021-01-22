package effect

import (
	"fmt"

	effectIntf "gotracker/internal/format/it/playback/effect/intf"
	"gotracker/internal/player/intf"
)

// FinePatternDelay defines an fine pattern delay effect
type FinePatternDelay uint8 // 'S6x'

// Start triggers on the first tick, but before the Tick() function is called
func (e FinePatternDelay) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	m := p.(effectIntf.IT)
	m.AddRowTicks(int(x))
}

func (e FinePatternDelay) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
