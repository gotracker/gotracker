package effect

import (
	"fmt"
	"gotracker/internal/format/s3m/channel"
	"gotracker/internal/player/intf"
)

// ExtraFinePortaDown defines an extra-fine portamento down effect
type ExtraFinePortaDown uint8 // 'EEx'

// PreStart triggers when the effect enters onto the channel state
func (e ExtraFinePortaDown) PreStart(cs intf.Channel, ss intf.Song) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e ExtraFinePortaDown) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()

	mem := cs.GetMemory().(*channel.Memory)
	xx := mem.LastNonZero(uint8(e))
	y := xx & 0x0F

	doPortaDown(cs, float32(y), 1)
}

// Tick is called on every tick
func (e ExtraFinePortaDown) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e ExtraFinePortaDown) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e ExtraFinePortaDown) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
