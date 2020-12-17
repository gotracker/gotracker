package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/channel"
	"gotracker/internal/player/intf"
)

// Tremolo defines a tremolo effect
type Tremolo uint8 // 'R'

// PreStart triggers when the effect enters onto the channel state
func (e Tremolo) PreStart(cs intf.Channel, ss intf.Song) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e Tremolo) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e Tremolo) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	xy := mem.LastNonZero(uint8(e))
	if currentTick == 0 {
		trem := cs.GetTremoloOscillator()
		trem.Pos = 0
	} else {
		x := xy >> 4
		y := xy & 0x0f
		doTremolo(cs, currentTick, x, y, 4)
	}
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e Tremolo) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e Tremolo) String() string {
	return fmt.Sprintf("R%0.2x", uint8(e))
}
