package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	effectIntf "gotracker/internal/format/s3m/playback/effect/intf"
	"gotracker/internal/player/intf"
)

// SetTempo defines a set tempo effect
type SetTempo uint8 // 'T'

// PreStart triggers when the effect enters onto the channel state
func (e SetTempo) PreStart(cs intf.Channel, p intf.Playback) {
	if e > 0x20 {
		m := p.(effectIntf.S3M)
		m.SetTempo(int(e))
	}
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetTempo) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
}

// Tick is called on every tick
func (e SetTempo) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	m := p.(effectIntf.S3M)
	switch uint8(e >> 4) {
	case 0: // decrease tempo
		if currentTick != 0 {
			mem := cs.GetMemory().(*channel.Memory)
			val := int(mem.TempoDecrease(uint8(e & 0x0F)))
			m.DecreaseTempo(val)
		}
	case 1: // increase tempo
		if currentTick != 0 {
			mem := cs.GetMemory().(*channel.Memory)
			val := int(mem.TempoIncrease(uint8(e & 0x0F)))
			m.IncreaseTempo(val)
		}
	default:
		m.SetTempo(int(e))
	}
}

func (e SetTempo) String() string {
	return fmt.Sprintf("T%0.2x", uint8(e))
}
