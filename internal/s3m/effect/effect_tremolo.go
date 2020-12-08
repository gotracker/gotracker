package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
)

type EffectTremolo uint8 // 'R'

func (e EffectTremolo) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectTremolo) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

func (e EffectTremolo) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	xy := cs.GetEffectSharedMemory(uint8(e))
	if currentTick == 0 {
		trem := cs.GetTremoloOscillator()
		trem.Pos = 0
	} else {
		x := xy >> 4
		y := xy & 0x0f
		doTremolo(cs, currentTick, x, y, 4)
	}
}

func (e EffectTremolo) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e EffectTremolo) String() string {
	return fmt.Sprintf("R%0.2x", uint8(e))
}
