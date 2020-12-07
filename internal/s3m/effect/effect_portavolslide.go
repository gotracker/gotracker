package effect

import "s3mplayer/internal/player/intf"

type EffectPortaVolumeSlide struct { // 'L'
	intf.CombinedEffect
}

func NewEffectPortaVolumeSlide(val uint8) EffectPortaVolumeSlide {
	pvs := EffectPortaVolumeSlide{}
	pvs.Effects = append(pvs.Effects, EffectVolumeSlide(val), EffectPortaToNote(0x00))
	return pvs
}
