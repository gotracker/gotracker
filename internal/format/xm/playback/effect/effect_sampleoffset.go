package effect

import (
	"fmt"

	"github.com/gotracker/gomixing/sampling"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// SampleOffset defines a sample offset effect
type SampleOffset uint8 // '9'

// Start triggers on the first tick, but before the Tick() function is called
func (e SampleOffset) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()
	mem := cs.GetMemory().(*channel.Memory)
	xx := mem.SampleOffset(uint8(e))
	cs.SetTargetPos(sampling.Pos{Pos: int(xx) * 0x100})
	return nil
}

func (e SampleOffset) String() string {
	return fmt.Sprintf("9%0.2x", uint8(e))
}
