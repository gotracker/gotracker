package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// FineVibrato defines an fine vibrato effect
type FineVibrato uint8 // 'U'

// PreStart triggers when the effect enters onto the channel state
func (e FineVibrato) PreStart(cs intf.Channel, p intf.Playback) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e FineVibrato) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
}

// Tick is called on every tick
func (e FineVibrato) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	xy := mem.Vibrato(uint8(e))
	if currentTick == 0 {
		vib := cs.GetVibratoOscillator()
		vib.Pos = 0
	} else {
		x := xy >> 4
		y := xy & 0x0f
		doVibrato(cs, currentTick, x, y, 1)
	}
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e FineVibrato) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e FineVibrato) String() string {
	return fmt.Sprintf("U%0.2x", uint8(e))
}
