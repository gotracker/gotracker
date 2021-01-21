package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// FinePortaUp defines an fine portamento up effect
type FinePortaUp uint8 // 'E1x'

// Start triggers on the first tick, but before the Tick() function is called
func (e FinePortaUp) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()

	mem := cs.GetMemory().(*channel.Memory)
	xy := mem.FinePortaUp(uint8(e))
	y := xy & 0x0F

	doPortaUp(cs, float32(y), 4, mem.LinearFreqSlides)
}

func (e FinePortaUp) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
