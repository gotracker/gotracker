package effect

import "s3mplayer/internal/player/intf"

type EffectFinePatternDelay uint8 // 'S6x'

func (e EffectFinePatternDelay) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectFinePatternDelay) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	ss.AddRowTicks(int(x))
}

func (e EffectFinePatternDelay) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

func (e EffectFinePatternDelay) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}
