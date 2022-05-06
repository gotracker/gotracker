package effect

import (
	"fmt"

	"github.com/gotracker/gotracker/internal/format/s3m/layout/channel"
	effectIntf "github.com/gotracker/gotracker/internal/format/s3m/playback/effect/intf"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// SetTempo defines a set tempo effect
type SetTempo ChannelCommand // 'T'

// PreStart triggers when the effect enters onto the channel state
func (e SetTempo) PreStart(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	if e > 0x20 {
		m := p.(effectIntf.S3M)
		if err := m.SetTempo(int(e)); err != nil {
			return err
		}
	}
	return nil
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetTempo) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Tick is called on every tick
func (e SetTempo) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	m := p.(effectIntf.S3M)
	switch channel.DataEffect(e >> 4) {
	case 0: // decrease tempo
		if currentTick != 0 {
			mem := cs.GetMemory()
			val := int(mem.TempoDecrease(channel.DataEffect(e & 0x0F)))
			if err := m.DecreaseTempo(val); err != nil {
				return err
			}
		}
	case 1: // increase tempo
		if currentTick != 0 {
			mem := cs.GetMemory()
			val := int(mem.TempoIncrease(channel.DataEffect(e & 0x0F)))
			if err := m.IncreaseTempo(val); err != nil {
				return err
			}
		}
	default:
		if err := m.SetTempo(int(e)); err != nil {
			return err
		}
	}
	return nil
}

func (e SetTempo) String() string {
	return fmt.Sprintf("T%0.2x", channel.DataEffect(e))
}
