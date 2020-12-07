package effect

import "s3mplayer/internal/player/intf"

type EffectVibrato uint8 // 'H'

func (e EffectVibrato) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectVibrato) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()
}

func (e EffectVibrato) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
	mem := cs.GetMemory()
	xy := mem.Vibrato(uint8(e))
	if currentTick == 0 {
		vib := cs.GetVibratoOscillator()
		vib.Pos = 0
	} else {
		x := xy >> 4
		y := xy & 0x0f
		doVibrato(cs, currentTick, x, y, 4)
	}
}

func (e EffectVibrato) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}
