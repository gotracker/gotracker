package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// FinePortaDown defines an fine portamento down effect
type FinePortaDown uint8 // 'EFx'

// PreStart triggers when the effect enters onto the channel state
func (e FinePortaDown) PreStart(cs intf.Channel, p intf.Playback) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e FinePortaDown) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()

	mem := cs.GetMemory().(*channel.Memory)
	xx := mem.LastNonZero(uint8(e))
	y := xx & 0x0F

	doPortaDown(cs, float32(y), 4)
}

// Tick is called on every tick
func (e FinePortaDown) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e FinePortaDown) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e FinePortaDown) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
