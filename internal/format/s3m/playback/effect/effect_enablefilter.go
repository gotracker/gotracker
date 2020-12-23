package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// EnableFilter defines a set filter enable effect
type EnableFilter uint8 // 'S0x'

// PreStart triggers when the effect enters onto the channel state
func (e EnableFilter) PreStart(cs intf.Channel, ss intf.Song) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e EnableFilter) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf
	on := x != 0

	// TODO: build lowpass filter, then enable/disable it!
	_ = on
}

// Tick is called on every tick
func (e EnableFilter) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e EnableFilter) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e EnableFilter) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
