package effect

import (
	"fmt"

	"github.com/gotracker/gotracker/internal/comparison"
	"github.com/gotracker/gotracker/internal/format/it/layout/channel"
	"github.com/gotracker/gotracker/internal/player/intf"
	"github.com/gotracker/gotracker/internal/song/note"
)

// PortaToNote defines a portamento-to-note effect
type PortaToNote channel.DataEffect // 'G'

// Start triggers on the first tick, but before the Tick() function is called
func (e PortaToNote) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
	if cmd := cs.GetData(); cmd != nil && cmd.HasNote() {
		cs.SetPortaTargetPeriod(cs.GetTargetPeriod())
		cs.SetNotePlayTick(false, note.ActionContinue, 0)
	}
	return nil
}

// Tick is called on every tick
func (e PortaToNote) Tick(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, currentTick int) error {
	mem := cs.GetMemory()
	xx := mem.PortaToNote(channel.DataEffect(e))

	// vibrato modifies current period for portamento
	period := cs.GetPeriod()
	if period == nil {
		return nil
	}
	period = period.AddDelta(cs.GetPeriodDelta()).(note.Period)
	ptp := cs.GetPortaTargetPeriod()
	if note.ComparePeriods(period, ptp) == comparison.SpaceshipRightGreater {
		return doPortaUpToNote(cs, float32(xx), 4, ptp, mem.Shared.LinearFreqSlides) // subtracts
	} else {
		return doPortaDownToNote(cs, float32(xx), 4, ptp, mem.Shared.LinearFreqSlides) // adds
	}
}

func (e PortaToNote) String() string {
	return fmt.Sprintf("G%0.2x", channel.DataEffect(e))
}
