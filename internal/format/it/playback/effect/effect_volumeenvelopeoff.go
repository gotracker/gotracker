package effect

import (
	"fmt"

	"github.com/gotracker/gotracker/internal/format/it/layout/channel"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// VolumeEnvelopeOff defines a volume envelope: off effect
type VolumeEnvelopeOff channel.DataEffect // 'S77'

// Start triggers on the first tick, but before the Tick() function is called
func (e VolumeEnvelopeOff) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	cs.SetVolumeEnvelopeEnable(false)
	return nil
}

func (e VolumeEnvelopeOff) String() string {
	return fmt.Sprintf("S%0.2x", channel.DataEffect(e))
}
