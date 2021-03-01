package effect

import (
	"fmt"

	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/intf"
)

// SetGlobalVolume defines a set global volume effect
type SetGlobalVolume uint8 // 'V'

// PreStart triggers when the effect enters onto the channel state
func (e SetGlobalVolume) PreStart(cs intf.Channel, p intf.Playback) error {
	v := volume.Volume(uint8(e)) / 0x80
	if v > 1 {
		v = 1
	}
	p.SetGlobalVolume(v)
	return nil
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetGlobalVolume) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

func (e SetGlobalVolume) String() string {
	return fmt.Sprintf("V%0.2x", uint8(e))
}
