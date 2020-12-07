package effect

import "s3mplayer/internal/player/intf"

type EffectRowJump uint8 // 'C'

func (e EffectRowJump) PreStart(cs intf.Channel, ss intf.Song) {
	ss.SetCurrentRow(uint8((e>>4)*10 + (e & 0x0f)))
}

func (e EffectRowJump) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
}

func (e EffectRowJump) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

func (e EffectRowJump) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}
