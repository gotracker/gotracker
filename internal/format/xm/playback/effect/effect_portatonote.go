package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// PortaToNote defines a portamento-to-note effect
type PortaToNote uint8 // '3'

// PreStart triggers when the effect enters onto the channel state
func (e PortaToNote) PreStart(cs intf.Channel, p intf.Playback) {
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
func (e PortaToNote) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
	cs.SetKeepFinetune(true)
}

// Tick is called on every tick
func (e PortaToNote) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	xx := mem.PortaToNote(uint8(e))

	period := cs.GetPeriod()
	ptp := cs.GetPortaTargetPeriod()
	if currentTick != 0 {
		if period > ptp {
			doPortaUpToNote(cs, float32(xx), 4, ptp, mem.LinearFreqSlides) // subtracts
		} else {
			doPortaDownToNote(cs, float32(xx), 4, ptp, mem.LinearFreqSlides) // adds
		}
	}
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e PortaToNote) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e PortaToNote) String() string {
	return fmt.Sprintf("3%0.2x", uint8(e))
}
