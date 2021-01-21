package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// PortaToNote defines a portamento-to-note effect
type PortaToNote uint8 // '3'

// Start triggers on the first tick, but before the Tick() function is called
func (e PortaToNote) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
	if cmd, ok := cs.GetData().(*channel.Data); ok && cmd.HasNote() {
		cs.SetPortaTargetPeriod(cs.GetTargetPeriod())
		cs.SetDoRetriggerNote(false)
	}
}

// Tick is called on every tick
func (e PortaToNote) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	xx := mem.PortaToNote(uint8(e))

	period := cs.GetPeriod()
	ptp := cs.GetPortaTargetPeriod()
	if currentTick != 0 {
		if note.ComparePeriods(period, ptp) == note.CompareRightHigher {
			doPortaUpToNote(cs, float32(xx), 4, ptp, mem.LinearFreqSlides) // subtracts
		} else {
			doPortaDownToNote(cs, float32(xx), 4, ptp, mem.LinearFreqSlides) // adds
		}
	}
}

func (e PortaToNote) String() string {
	return fmt.Sprintf("3%0.2x", uint8(e))
}
