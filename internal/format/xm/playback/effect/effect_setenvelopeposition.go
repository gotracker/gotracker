package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// SetEnvelopePosition defines a set envelope position effect
type SetEnvelopePosition uint8 // 'Lxx'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetEnvelopePosition) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()

	xx := uint8(e)

	cs.SetEnvelopePosition(int(xx))
	return nil
}

func (e SetEnvelopePosition) String() string {
	return fmt.Sprintf("L%0.2x", uint8(e))
}
