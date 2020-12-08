package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
)

type EffectSetPanPosition uint8 // 'S8x'

func (e EffectSetPanPosition) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectSetPanPosition) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	cs.SetPan(x)
}

func (e EffectSetPanPosition) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

func (e EffectSetPanPosition) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e EffectSetPanPosition) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
