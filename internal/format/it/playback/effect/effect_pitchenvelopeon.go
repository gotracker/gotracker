package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// PitchEnvelopeOn defines a panning envelope: on effect
type PitchEnvelopeOn channel.DataEffect // 'S7C'

// Start triggers on the first tick, but before the Tick() function is called
func (e PitchEnvelopeOn) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	cs.SetPitchEnvelopeEnable(true)
	return nil
}

func (e PitchEnvelopeOn) String() string {
	return fmt.Sprintf("S%0.2x", channel.DataEffect(e))
}
