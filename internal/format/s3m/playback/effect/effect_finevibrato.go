package effect

import (
	"fmt"

	"github.com/gotracker/gotracker/internal/format/s3m/layout/channel"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// FineVibrato defines an fine vibrato effect
type FineVibrato ChannelCommand // 'U'

// Start triggers on the first tick, but before the Tick() function is called
func (e FineVibrato) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
	return nil
}

// Tick is called on every tick
func (e FineVibrato) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	mem := cs.GetMemory()
	x, y := mem.Vibrato(channel.DataEffect(e))
	// NOTE: JBC - S3M does not update on tick 0, but MOD does.
	if currentTick != 0 || mem.Shared.ModCompatibility {
		return doVibrato(cs, currentTick, x, y, 1)
	}
	return nil
}

func (e FineVibrato) String() string {
	return fmt.Sprintf("U%0.2x", channel.DataEffect(e))
}
