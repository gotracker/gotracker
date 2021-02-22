package effect

import (
	"fmt"

	"gotracker/internal/comparison"
	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// PortaToNote defines a portamento-to-note effect
type PortaToNote uint8 // 'G'

// Start triggers on the first tick, but before the Tick() function is called
func (e PortaToNote) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
	if cmd, ok := cs.GetData().(*channel.Data); ok && cmd.HasNote() {
		cs.SetPortaTargetPeriod(cs.GetTargetPeriod())
		cs.SetNotePlayTick(false, 0)
	}
}

// Tick is called on every tick
func (e PortaToNote) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	xx := mem.PortaToNote(uint8(e))

	// vibrato modifies current period for portamento
	period := cs.GetPeriod()
	if period == nil {
		return
	}
	period = period.AddDelta(cs.GetPeriodDelta()).(note.Period)
	ptp := cs.GetPortaTargetPeriod()
	if currentTick != 0 {
		if note.ComparePeriods(period, ptp) == comparison.SpaceshipRightGreater {
			doPortaUpToNote(cs, float32(xx), 4, ptp, mem.LinearFreqSlides) // subtracts
		} else {
			doPortaDownToNote(cs, float32(xx), 4, ptp, mem.LinearFreqSlides) // adds
		}
	}
}

func (e PortaToNote) String() string {
	return fmt.Sprintf("G%0.2x", uint8(e))
}
