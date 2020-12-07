package effect

import "s3mplayer/internal/player/intf"

type EffectStereoControl uint8 // 'SAx'

func (e EffectStereoControl) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectStereoControl) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	if x > 7 {
		cs.SetPan(x - 8)
	} else {
		cs.SetPan(x + 8)
	}
}

func (e EffectStereoControl) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

func (e EffectStereoControl) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}
