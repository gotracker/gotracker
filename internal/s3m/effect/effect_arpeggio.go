package effect

import "gotracker/internal/player/intf"

type EffectArpeggio uint8 // 'J'

func (e EffectArpeggio) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectArpeggio) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
}

func (e EffectArpeggio) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	xy := cs.GetEffectSharedMemory(uint8(e))
	x := (xy >> 4) - 8
	y := (xy & 0x0f) - 8
	doArpeggio(cs, currentTick, x, y)
}

func (e EffectArpeggio) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}
