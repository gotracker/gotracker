package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// VolumeEnvelopeOff defines a volume envelope: off effect
type VolumeEnvelopeOff uint8 // 'S77'

// Start triggers on the first tick, but before the Tick() function is called
func (e VolumeEnvelopeOff) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	cs.SetVolumeEnvelopeEnable(false)
}

func (e VolumeEnvelopeOff) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
