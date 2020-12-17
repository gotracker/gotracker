package effect

import (
	"fmt"

	"github.com/heucuva/gomixing/sampling"

	"gotracker/internal/format/s3m/channel"
	"gotracker/internal/player/intf"
)

// SampleOffset defines a sample offset effect
type SampleOffset uint8 // 'O'

// PreStart triggers when the effect enters onto the channel state
func (e SampleOffset) PreStart(cs intf.Channel, ss intf.Song) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SampleOffset) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
	mem := cs.GetMemory().(*channel.Memory)
	xx := mem.SampleOffset(uint8(e))
	cs.SetTargetPos(sampling.Pos{Pos: int(xx) * 0x100})
}

// Tick is called on every tick
func (e SampleOffset) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e SampleOffset) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e SampleOffset) String() string {
	return fmt.Sprintf("O%0.2x", uint8(e))
}
