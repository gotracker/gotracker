package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
)

// PortaVolumeSlide defines a portamento-to-note combined with a volume slide effect
type PortaVolumeSlide struct { // 'L'
	intf.CombinedEffect
}

// NewPortaVolumeSlide creates a new PortaVolumeSlide object
func NewPortaVolumeSlide(val uint8) PortaVolumeSlide {
	pvs := PortaVolumeSlide{}
	pvs.Effects = append(pvs.Effects, VolumeSlide(val), PortaToNote(0x00))
	return pvs
}

func (e PortaVolumeSlide) String() string {
	return fmt.Sprintf("L%0.2x", uint8(e.Effects[0].(VolumeSlide)))
}
