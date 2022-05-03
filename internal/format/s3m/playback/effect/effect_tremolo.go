package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// Tremolo defines a tremolo effect
type Tremolo ChannelCommand // 'R'

// Start triggers on the first tick, but before the Tick() function is called
func (e Tremolo) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e Tremolo) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	mem := cs.GetMemory()
	x, y := mem.Tremolo(channel.DataEffect(e))
	// NOTE: JBC - S3M does not update on tick 0, but MOD does.
	// Maybe need to add a flag for converted MOD backward compatibility?
	if currentTick != 0 {
		return doTremolo(cs, currentTick, channel.DataEffect(x), channel.DataEffect(y), 4)
	}
	return nil
}

func (e Tremolo) String() string {
	return fmt.Sprintf("R%0.2x", channel.DataEffect(e))
}
