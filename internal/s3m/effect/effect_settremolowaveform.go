package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/oscillator"
)

type EffectSetTremoloWaveform uint8 // 'S4x'

func (e EffectSetTremoloWaveform) PreStart(cs intf.Channel, ss intf.Song) {
}

func (e EffectSetTremoloWaveform) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	trem := cs.GetTremoloOscillator()
	trem.Table = oscillator.WaveTableSelect(x)
}

func (e EffectSetTremoloWaveform) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

func (e EffectSetTremoloWaveform) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e EffectSetTremoloWaveform) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
