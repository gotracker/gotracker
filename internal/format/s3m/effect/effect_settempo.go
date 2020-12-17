package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/channel"
	"gotracker/internal/player/intf"
)

// SetTempo defines a set tempo effect
type SetTempo uint8 // 'T'

// PreStart triggers when the effect enters onto the channel state
func (e SetTempo) PreStart(cs intf.Channel, ss intf.Song) {
	if e > 0x20 {
		ss.SetTempo(int(e))
	}
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetTempo) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e SetTempo) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	switch uint8(e >> 4) {
	case 0: // decrease tempo
		if currentTick != 0 {
			mem := cs.GetMemory().(*channel.Memory)
			val := int(mem.TempoDecrease(uint8(e & 0x0F)))
			ss.DecreaseTempo(val)
		}
	case 1: // increase tempo
		if currentTick != 0 {
			mem := cs.GetMemory().(*channel.Memory)
			val := int(mem.TempoIncrease(uint8(e & 0x0F)))
			ss.IncreaseTempo(val)
		}
	default:
	}
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e SetTempo) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e SetTempo) String() string {
	return fmt.Sprintf("T%0.2x", uint8(e))
}
