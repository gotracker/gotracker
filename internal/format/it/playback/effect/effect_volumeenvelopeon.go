package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// VolumeEnvelopeOn defines a volume envelope: on effect
type VolumeEnvelopeOn uint8 // 'S78'

// Start triggers on the first tick, but before the Tick() function is called
func (e VolumeEnvelopeOn) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	cs.SetVolumeEnvelopeEnable(true)
}

func (e VolumeEnvelopeOn) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
