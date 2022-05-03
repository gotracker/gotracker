package effect

import (
	"fmt"

	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// SetGlobalVolume defines a set global volume effect
type SetGlobalVolume channel.DataEffect // 'V'

// PreStart triggers when the effect enters onto the channel state
func (e SetGlobalVolume) PreStart(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	v := volume.Volume(channel.DataEffect(e)) / 0x80
	if v > 1 {
		v = 1
	}
	p.SetGlobalVolume(v)
	return nil
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetGlobalVolume) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

func (e SetGlobalVolume) String() string {
	return fmt.Sprintf("V%0.2x", channel.DataEffect(e))
}
