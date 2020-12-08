package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
)

// FinePatternDelay defines an fine pattern delay effect
type FinePatternDelay uint8 // 'S6x'

// PreStart triggers when the effect enters onto the channel state
func (e FinePatternDelay) PreStart(cs intf.Channel, ss intf.Song) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e FinePatternDelay) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	ss.AddRowTicks(int(x))
}

// Tick is called on every tick
func (e FinePatternDelay) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e FinePatternDelay) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e FinePatternDelay) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
