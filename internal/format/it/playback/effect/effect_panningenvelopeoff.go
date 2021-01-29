package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// PanningEnvelopeOff defines a panning envelope: off effect
type PanningEnvelopeOff uint8 // 'S79'

// Start triggers on the first tick, but before the Tick() function is called
func (e PanningEnvelopeOff) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	cs.SetPanningEnvelopeEnable(false)
}

func (e PanningEnvelopeOff) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
