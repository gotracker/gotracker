package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// SetFinetune defines a mod-style set finetune effect
type SetFinetune uint8 // 'E5x'

// PreStart triggers when the effect enters onto the channel state
func (e SetFinetune) PreStart(cs intf.Channel, p intf.Playback) {
	x := uint8(e) & 0xf

	inst := cs.GetTargetInst()
	if inst != nil {
		inst.SetFinetune(int8(x) - 8)
	}
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetFinetune) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e SetFinetune) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e SetFinetune) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e SetFinetune) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}