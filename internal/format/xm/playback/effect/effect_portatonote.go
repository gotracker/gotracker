package effect

import (
	"fmt"

	"gotracker/internal/comparison"
	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// PortaToNote defines a portamento-to-note effect
type PortaToNote uint8 // '3'

// Start triggers on the first tick, but before the Tick() function is called
func (e PortaToNote) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
	if cmd, ok := cs.GetData().(*channel.Data); ok && cmd.HasNote() {
		cs.SetPortaTargetPeriod(cs.GetTargetPeriod())
		cs.SetNotePlayTick(false, 0)
	}
	return nil
}

// Tick is called on every tick
func (e PortaToNote) Tick(cs intf.Channel, p intf.Playback, currentTick int) error {
	mem := cs.GetMemory().(*channel.Memory)
	xx := mem.PortaToNote(uint8(e))

	period := cs.GetPeriod()
	ptp := cs.GetPortaTargetPeriod()
	if currentTick != 0 {
		if note.ComparePeriods(period, ptp) == comparison.SpaceshipRightGreater {
			return doPortaUpToNote(cs, float32(xx), 4, ptp, mem.LinearFreqSlides) // subtracts
		} else {
			return doPortaDownToNote(cs, float32(xx), 4, ptp, mem.LinearFreqSlides) // adds
		}
	}
	return nil
}

func (e PortaToNote) String() string {
	return fmt.Sprintf("3%0.2x", uint8(e))
}
