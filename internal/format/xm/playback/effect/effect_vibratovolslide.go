package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// VibratoVolumeSlide defines a combination vibrato and volume slide effect
type VibratoVolumeSlide struct { // '6'
	intf.CombinedEffect[channel.Memory, channel.Data]
}

// NewVibratoVolumeSlide creates a new VibratoVolumeSlide object
func NewVibratoVolumeSlide(val channel.DataEffect) VibratoVolumeSlide {
	vvs := VibratoVolumeSlide{}
	vvs.Effects = append(vvs.Effects, VolumeSlide(val), Vibrato(0x00))
	return vvs
}

func (e VibratoVolumeSlide) String() string {
	return fmt.Sprintf("6%0.2x", channel.DataEffect(e.Effects[0].(VolumeSlide)))
}
