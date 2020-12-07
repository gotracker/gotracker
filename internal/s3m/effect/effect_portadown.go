package effect

import "gotracker/internal/player/intf"

type EffectPortaDown uint8 // 'E'

func (e EffectPortaDown) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectPortaDown) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
}

func (e EffectPortaDown) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	xx := cs.GetEffectSharedMemory(uint8(e))
	x := xx >> 4
	y := xx & 0x0F

	if x == 0x0F { // fine portamento down
		if currentTick == 0 {
			doPortaDown(cs, float32(y), 4)
		}
	} else if x == 0x0E { // extra-fine portamento down
		if currentTick == 0 {
			doPortaDown(cs, float32(y), 1)
		}
	} else {
		if currentTick != 0 {
			doPortaDown(cs, float32(xx), 4)
		}
	}
}

func (e EffectPortaDown) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}
