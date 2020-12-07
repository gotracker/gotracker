package effect

import "gotracker/internal/player/intf"

type EffectTremor uint8 // 'I'

func (e EffectTremor) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectTremor) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

func (e EffectTremor) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	xy := cs.GetEffectSharedMemory(uint8(e))
	x := int((xy >> 4) + 1)
	y := int((xy & 0x0f) + 1)
	doTremor(cs, currentTick, x, y)
}

func (e EffectTremor) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}
