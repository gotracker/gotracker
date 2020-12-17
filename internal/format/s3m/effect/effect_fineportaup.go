package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/channel"
	"gotracker/internal/module/player/intf"
)

// FinePortaUp defines an fine portamento up effect
type FinePortaUp uint8 // 'FFx'

// PreStart triggers when the effect enters onto the channel state
func (e FinePortaUp) PreStart(cs intf.Channel, ss intf.Song) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e FinePortaUp) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()

	mem := cs.GetMemory().(*channel.Memory)
	xx := mem.LastNonZero(uint8(e))
	y := xx & 0x0F

	doPortaUp(cs, float32(y), 4)
}

// Tick is called on every tick
func (e FinePortaUp) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e FinePortaUp) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e FinePortaUp) String() string {
	return fmt.Sprintf("F%0.2x", uint8(e))
}
