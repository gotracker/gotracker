package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// ExtraFinePortaUp defines an extra-fine portamento up effect
type ExtraFinePortaUp uint8 // 'FEx'

// Start triggers on the first tick, but before the Tick() function is called
func (e ExtraFinePortaUp) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()

	y := uint8(e) & 0x0F

	doPortaUp(cs, float32(y), 1)
}

func (e ExtraFinePortaUp) String() string {
	return fmt.Sprintf("F%0.2x", uint8(e))
}
