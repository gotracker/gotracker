package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// FinePortaDown defines an fine portamento down effect
type FinePortaDown uint8 // 'E2x'

// Start triggers on the first tick, but before the Tick() function is called
func (e FinePortaDown) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()

	mem := cs.GetMemory().(*channel.Memory)
	xy := mem.FinePortaDown(uint8(e))
	y := xy & 0x0F

	doPortaDown(cs, float32(y), 4, mem.LinearFreqSlides)
}

func (e FinePortaDown) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
