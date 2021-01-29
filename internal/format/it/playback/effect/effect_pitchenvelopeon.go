package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// PitchEnvelopeOn defines a panning envelope: on effect
type PitchEnvelopeOn uint8 // 'S7C'

// Start triggers on the first tick, but before the Tick() function is called
func (e PitchEnvelopeOn) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	cs.SetPitchEnvelopeEnable(true)
}

func (e PitchEnvelopeOn) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
