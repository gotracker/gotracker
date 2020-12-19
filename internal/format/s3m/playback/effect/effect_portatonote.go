package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// PortaToNote defines a portamento-to-note effect
type PortaToNote uint8 // 'G'

// PreStart triggers when the effect enters onto the channel state
func (e PortaToNote) PreStart(cs intf.Channel, ss intf.Song) {
	cmd := cs.GetData().(*channel.Data)
	if cmd == nil {
		return
	}

	if cmd.What.HasNote() {
		cs.SetPortaTargetPeriod(cs.GetTargetPeriod())
		cs.SetDoRetriggerNote(false)
	}
}

// Start triggers on the first tick, but before the Tick() function is called
func (e PortaToNote) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
}

// Tick is called on every tick
func (e PortaToNote) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	xx := mem.PortaToNote(uint8(e))

	period := cs.GetPeriod()
	ptp := cs.GetPortaTargetPeriod()
	if currentTick != 0 {
		if period > ptp {
			doPortaUpToNote(cs, float32(xx), 4, ptp) // subtracts
		} else {
			doPortaDownToNote(cs, float32(xx), 4, ptp) // adds
		}
	}
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e PortaToNote) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e PortaToNote) String() string {
	return fmt.Sprintf("G%0.2x", uint8(e))
}
