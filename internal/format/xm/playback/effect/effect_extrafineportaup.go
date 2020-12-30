package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// ExtraFinePortaUp defines an extra-fine portamento up effect
type ExtraFinePortaUp uint8 // 'X1x'

// PreStart triggers when the effect enters onto the channel state
func (e ExtraFinePortaUp) PreStart(cs intf.Channel, p intf.Playback) {
	cs.SetKeepFinetune(true)
}

// Start triggers on the first tick, but before the Tick() function is called
func (e ExtraFinePortaUp) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()

	mem := cs.GetMemory().(*channel.Memory)
	xx := mem.ExtraFinePortaUp(uint8(e))
	y := xx & 0x0F

	doPortaUp(cs, float32(y), 1, mem.LinearFreqSlides)
}

// Tick is called on every tick
func (e ExtraFinePortaUp) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e ExtraFinePortaUp) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e ExtraFinePortaUp) String() string {
	return fmt.Sprintf("F%0.2x", uint8(e))
}
