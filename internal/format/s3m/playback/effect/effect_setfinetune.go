package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
	"gotracker/internal/song/note"
)

// SetFinetune defines a mod-style set finetune effect
type SetFinetune ChannelCommand // 'S2x'

// PreStart triggers when the effect enters onto the channel state
func (e SetFinetune) PreStart(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	x := channel.DataEffect(e) & 0xf

	inst := cs.GetTargetInst()
	if inst != nil {
		ft := (note.Finetune(x) - 8) * 4
		inst.SetFinetune(ft)
	}
	return nil
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetFinetune) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

func (e SetFinetune) String() string {
	return fmt.Sprintf("S%0.2x", channel.DataEffect(e))
}
