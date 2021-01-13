package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// FinePortaDown defines an fine portamento down effect
type FinePortaDown uint8 // 'EFx'

// Start triggers on the first tick, but before the Tick() function is called
func (e FinePortaDown) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()

	y := uint8(e) & 0x0F

	doPortaDown(cs, float32(y), 4)
}

func (e FinePortaDown) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
