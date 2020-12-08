package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
)

type EffectOrderJump uint8 // 'B'

func (e EffectOrderJump) PreStart(cs intf.Channel, ss intf.Song) {
	ss.SetCurrentOrder(uint8(e))
}

func (e EffectOrderJump) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

func (e EffectOrderJump) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

func (e EffectOrderJump) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e EffectOrderJump) String() string {
	return fmt.Sprintf("B%0.2x", uint8(e))
}
