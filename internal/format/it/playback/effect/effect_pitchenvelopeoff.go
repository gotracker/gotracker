package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// PitchEnvelopeOff defines a panning envelope: off effect
type PitchEnvelopeOff uint8 // 'S7B'

// Start triggers on the first tick, but before the Tick() function is called
func (e PitchEnvelopeOff) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	cs.SetPitchEnvelopeEnable(false)
}

func (e PitchEnvelopeOff) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
