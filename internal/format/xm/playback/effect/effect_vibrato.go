package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// Vibrato defines a vibrato effect
type Vibrato channel.DataEffect // '4'

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
	// NOTE: JBC - XM updates on tick 0, but MOD does not.
	// Just have to eat this incompatibility, I guess...
	return doVibrato(cs, currentTick, x, y, 4)
}

func (e Vibrato) String() string {
	return fmt.Sprintf("4%0.2x", channel.DataEffect(e))
}
