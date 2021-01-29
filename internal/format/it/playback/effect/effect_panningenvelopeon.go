package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// PanningEnvelopeOn defines a panning envelope: on effect
type PanningEnvelopeOn uint8 // 'S7A'

// Start triggers on the first tick, but before the Tick() function is called
func (e PanningEnvelopeOn) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	cs.SetPanningEnvelopeEnable(true)
}

func (e PanningEnvelopeOn) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
