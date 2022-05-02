package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// PanningEnvelopeOn defines a panning envelope: on effect
type PanningEnvelopeOn uint8 // 'S7A'

// Start triggers on the first tick, but before the Tick() function is called
func (e PanningEnvelopeOn) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	cs.SetPanningEnvelopeEnable(true)
	return nil
}

func (e PanningEnvelopeOn) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
