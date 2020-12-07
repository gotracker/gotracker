package effect

import "gotracker/internal/player/intf"

type EffectPortaUp uint8 // 'F'

func (e EffectPortaUp) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectPortaUp) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
}

func (e EffectPortaUp) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	xx := cs.GetEffectSharedMemory(uint8(e))
	x := xx >> 4
	y := xx & 0x0F

	if x == 0x0F { // fine portamento up
		if currentTick == 0 {
			doPortaUp(cs, float32(y), 4)
		}
	} else if x == 0x0E { // extra-fine portamento up
		if currentTick == 0 {
			doPortaUp(cs, float32(y), 1)
		}
	} else {
		if currentTick != 0 {
			doPortaUp(cs, float32(xx), 4)
		}
	}
}

func (e EffectPortaUp) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}
