package effect

import (
	"s3mplayer/internal/player/intf"
	"s3mplayer/internal/player/oscillator"
)

type EffectSetVibratoWaveform uint8 // 'S3x'

func (e EffectSetVibratoWaveform) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectSetVibratoWaveform) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	vib := cs.GetVibratoOscillator()
	vib.Table = oscillator.WaveTableSelect(x)
}

func (e EffectSetVibratoWaveform) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

func (e EffectSetVibratoWaveform) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}
