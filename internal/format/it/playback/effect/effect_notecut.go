package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// NoteCut defines a note cut effect
type NoteCut channel.DataEffect // 'SCx'

// Start triggers on the first tick, but before the Tick() function is called
func (e NoteCut) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e NoteCut) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	x := channel.DataEffect(e) & 0xf

	if x != 0 && currentTick == int(x) {
		cs.FreezePlayback()
	}
	return nil
}

func (e NoteCut) String() string {
	return fmt.Sprintf("S%0.2x", channel.DataEffect(e))
}
