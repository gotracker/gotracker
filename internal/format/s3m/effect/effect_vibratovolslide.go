package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// VibratoVolumeSlide defines a combination vibrato and volume slide effect
type VibratoVolumeSlide struct { // 'K'
	intf.CombinedEffect
}

// NewVibratoVolumeSlide creates a new VibratoVolumeSlide object
func NewVibratoVolumeSlide(val uint8) VibratoVolumeSlide {
	vvs := VibratoVolumeSlide{}
	vvs.Effects = append(vvs.Effects, VolumeSlide(val), Vibrato(0x00))
	return vvs
}

func (e VibratoVolumeSlide) String() string {
	return fmt.Sprintf("K%0.2x", uint8(e.Effects[0].(VolumeSlide)))
}
