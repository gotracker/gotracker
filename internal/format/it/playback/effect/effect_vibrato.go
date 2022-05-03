package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// Vibrato defines a vibrato effect
type Vibrato channel.DataEffect // 'H'

// Start triggers on the first tick, but before the Tick() function is called
func (e Vibrato) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
	return nil
}

// Tick is called on every tick
func (e Vibrato) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	mem := cs.GetMemory()
	x, y := mem.Vibrato(channel.DataEffect(e))
	if mem.OldEffectMode {
		if currentTick != 0 {
			return doVibrato(cs, currentTick, x, y, 8)
		}
	} else {
		return doVibrato(cs, currentTick, x, y, 4)
	}
	return nil
}

func (e Vibrato) String() string {
	return fmt.Sprintf("H%0.2x", channel.DataEffect(e))
}
