package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// NoteCut defines a note cut effect
type NoteCut uint8 // 'SCx'

// PreStart triggers when the effect enters onto the channel state
func (e NoteCut) PreStart(cs intf.Channel, ss intf.Song) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e NoteCut) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e NoteCut) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	x := uint8(e) & 0xf

	if x != 0 && currentTick == int(x) {
		cs.FreezePlayback()
	}
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e NoteCut) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e NoteCut) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
