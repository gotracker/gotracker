package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	effectIntf "gotracker/internal/format/xm/playback/effect/intf"
	"gotracker/internal/player/intf"
)

// SetTempo defines a set tempo effect
type SetTempo uint8 // 'F'

// PreStart triggers when the effect enters onto the channel state
func (e SetTempo) PreStart(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	if e > 0x20 {
		m := p.(effectIntf.XM)
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
	m := p.(effectIntf.XM)
	return m.SetTempo(int(e))
}

func (e SetTempo) String() string {
	return fmt.Sprintf("F%0.2x", uint8(e))
}
