package effect

import "gotracker/internal/player/intf"

type EffectVolumeSlide uint8 // 'D'

func (e EffectVolumeSlide) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectVolumeSlide) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

func (e EffectVolumeSlide) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	v := cs.GetEffectSharedMemory(uint8(e))
	x := uint8(v >> 4)
	y := uint8(v & 0x0F)

	if x == 0 { // decrease every tick
		if y == 0x0F {
			doVolSlide(cs, -float32(y), 1.0)
		} else if currentTick != 0 {
			doVolSlide(cs, -float32(y), 1.0)
		}
	} else if y == 0 { // increase every tick
		if currentTick != 0 {
			doVolSlide(cs, float32(x), 1.0)
		}
	} else if x == 0x0F { // finely decrease on the first tick
		if y != 0x0F && currentTick == 0 {
			doVolSlide(cs, -float32(y), 1.0)
		}
	} else if y == 0x0F { // finely increase on the first tick
		if x != 0x0F && currentTick == 0 {
			doVolSlide(cs, float32(x), 1.0)
		}
	}
}

func (e EffectVolumeSlide) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}
