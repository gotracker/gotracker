package effect

import (
	"fmt"
	"gotracker/internal/format/s3m/channel"
	"gotracker/internal/module/player/intf"
)

// Tremor defines a tremor effect
type Tremor uint8 // 'I'

// PreStart triggers when the effect enters onto the channel state
func (e Tremor) PreStart(cs intf.Channel, ss intf.Song) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e Tremor) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e Tremor) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	xy := mem.LastNonZero(uint8(e))
	x := int((xy >> 4) + 1)
	y := int((xy & 0x0f) + 1)
	doTremor(cs, currentTick, x, y)
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e Tremor) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e Tremor) String() string {
	return fmt.Sprintf("I%0.2x", uint8(e))
}
