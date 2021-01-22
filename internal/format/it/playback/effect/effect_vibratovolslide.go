package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
)

// VibratoVolumeSlide defines a combination vibrato and volume slide effect
type VibratoVolumeSlide struct { // 'K'
	intf.CombinedEffect
}

// NewVibratoVolumeSlide creates a new VibratoVolumeSlide object
func NewVibratoVolumeSlide(mem *channel.Memory, cd uint8, val uint8) VibratoVolumeSlide {
	vvs := VibratoVolumeSlide{}
	vs := volumeSlideFactory(mem, cd, val)
	vvs.Effects = append(vvs.Effects, vs, Vibrato(0x00))
	return vvs
}

func (e VibratoVolumeSlide) String() string {
	return fmt.Sprintf("K%0.2x", e.Effects[0].(uint8))
}
