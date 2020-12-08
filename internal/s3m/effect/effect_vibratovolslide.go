package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
)

type EffectVibratoVolumeSlide struct { // 'K'
	intf.CombinedEffect
}

func NewEffectVibratoVolumeSlide(val uint8) EffectVibratoVolumeSlide {
	vvs := EffectVibratoVolumeSlide{}
	vvs.Effects = append(vvs.Effects, EffectVolumeSlide(val), EffectVibrato(0x00))
	return vvs
}

func (e EffectVibratoVolumeSlide) String() string {
	return fmt.Sprintf("K%0.2x", uint8(e.Effects[0].(EffectVolumeSlide)))
}
