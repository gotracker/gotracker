package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// PitchEnvelopeOff defines a panning envelope: off effect
type PitchEnvelopeOff channel.DataEffect // 'S7B'

// Start triggers on the first tick, but before the Tick() function is called
func (e PitchEnvelopeOff) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	cs.SetPitchEnvelopeEnable(false)
	return nil
}

func (e PitchEnvelopeOff) String() string {
	return fmt.Sprintf("S%0.2x", channel.DataEffect(e))
}
