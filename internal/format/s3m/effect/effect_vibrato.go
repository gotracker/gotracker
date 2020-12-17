package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/channel"
	"gotracker/internal/module/player/intf"
)

// Vibrato defines a vibrato effect
type Vibrato uint8 // 'H'

// PreStart triggers when the effect enters onto the channel state
func (e Vibrato) PreStart(cs intf.Channel, ss intf.Song) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e Vibrato) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
}

// Tick is called on every tick
func (e Vibrato) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	mem := cs.GetMemory().(*channel.Memory)
	xy := mem.Vibrato(uint8(e))
	if currentTick == 0 {
		vib := cs.GetVibratoOscillator()
		vib.Pos = 0
	} else {
		x := xy >> 4
		y := xy & 0x0f
		doVibrato(cs, currentTick, x, y, 4)
	}
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e Vibrato) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e Vibrato) String() string {
	return fmt.Sprintf("H%0.2x", uint8(e))
}
