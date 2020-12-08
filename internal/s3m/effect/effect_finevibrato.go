package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
)

type EffectFineVibrato uint8 // 'U'

func (e EffectFineVibrato) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectFineVibrato) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
}

func (e EffectFineVibrato) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	mem := cs.GetMemory()
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

func (e EffectFineVibrato) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e EffectFineVibrato) String() string {
	return fmt.Sprintf("U%0.2x", uint8(e))
}
