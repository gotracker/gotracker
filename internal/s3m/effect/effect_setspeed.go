package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
)

type EffectSetSpeed uint8 // 'A'

func (e EffectSetSpeed) PreStart(cs intf.Channel, ss intf.Song) {
	if e != 0 {
		ss.SetTicks(int(e))
	}
}

func (e EffectSetSpeed) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

func (e EffectSetSpeed) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

func (e EffectSetSpeed) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e EffectSetSpeed) String() string {
	return fmt.Sprintf("A%0.2x", uint8(e))
}
