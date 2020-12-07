package effect

import "s3mplayer/internal/player/intf"

type EffectVibratoVolumeSlide struct { // 'K'
	intf.CombinedEffect
}

func NewEffectVibratoVolumeSlide(val uint8) EffectVibratoVolumeSlide {
	vvs := EffectVibratoVolumeSlide{}
	vvs.Effects = append(vvs.Effects, EffectVolumeSlide(val), EffectVibrato(0x00))
	return vvs
}
