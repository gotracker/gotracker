package effect

import (
	"fmt"

	"github.com/gotracker/gomixing/sampling"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// SampleOffset defines a sample offset effect
type SampleOffset uint8 // 'O'

// Start triggers on the first tick, but before the Tick() function is called
func (e SampleOffset) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	mem := cs.GetMemory()
	xx := mem.SampleOffset(uint8(e))

	pos := sampling.Pos{Pos: mem.HighOffset + int(xx)*0x100}
	if mem.OldEffectMode {
		if inst := cs.GetInstrument(); inst != nil && inst.GetLength().Pos < pos.Pos {
			cs.SetTargetPos(pos)
		}
	} else {
		cs.SetTargetPos(pos)
	}
	return nil
}

func (e SampleOffset) String() string {
	return fmt.Sprintf("O%0.2x", uint8(e))
}
