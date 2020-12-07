package effect

import "gotracker/internal/player/intf"

type EffectPatternDelay uint8 // 'SEx'

func (e EffectPatternDelay) PreStart(cs intf.Channel, ss intf.Song) {
	ss.SetPatternDelay(int(uint8(e) & 0x0F))
}

func (e EffectPatternDelay) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

func (e EffectPatternDelay) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

func (e EffectPatternDelay) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}
