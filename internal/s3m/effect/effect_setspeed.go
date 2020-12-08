package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
)

// SetSpeed defines a set speed effect
type SetSpeed uint8 // 'A'

// PreStart triggers when the effect enters onto the channel state
func (e SetSpeed) PreStart(cs intf.Channel, ss intf.Song) {
	if e != 0 {
		ss.SetTicks(int(e))
	}
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetSpeed) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e SetSpeed) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e SetSpeed) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e SetSpeed) String() string {
	return fmt.Sprintf("A%0.2x", uint8(e))
}
