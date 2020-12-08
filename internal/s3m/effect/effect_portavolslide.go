package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
)

type EffectPortaVolumeSlide struct { // 'L'
	intf.CombinedEffect
}

func NewEffectPortaVolumeSlide(val uint8) EffectPortaVolumeSlide {
	pvs := EffectPortaVolumeSlide{}
	pvs.Effects = append(pvs.Effects, EffectVolumeSlide(val), EffectPortaToNote(0x00))
	return pvs
}

func (e EffectPortaVolumeSlide) String() string {
	return fmt.Sprintf("L%0.2x", uint8(e.Effects[0].(EffectVolumeSlide)))
}
