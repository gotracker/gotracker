package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
)

type EffectNoteCut uint8 // 'SCx'

func (e EffectNoteCut) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectNoteCut) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

func (e EffectNoteCut) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	x := uint8(e) & 0xf

	if x != 0 && currentTick == int(x) {
		cs.FreezePlayback()
	}
}

func (e EffectNoteCut) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e EffectNoteCut) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
