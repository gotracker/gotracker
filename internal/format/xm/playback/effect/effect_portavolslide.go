package effect

import (
	"fmt"

	"github.com/gotracker/gotracker/internal/format/xm/layout/channel"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// PortaVolumeSlide defines a portamento-to-note combined with a volume slide effect
type PortaVolumeSlide struct { // '5'
	intf.CombinedEffect[channel.Memory, channel.Data]
}

// NewPortaVolumeSlide creates a new PortaVolumeSlide object
func NewPortaVolumeSlide(val channel.DataEffect) PortaVolumeSlide {
	pvs := PortaVolumeSlide{}
	pvs.Effects = append(pvs.Effects, VolumeSlide(val), PortaToNote(0x00))
	return pvs
}

func (e PortaVolumeSlide) String() string {
	return fmt.Sprintf("5%0.2x", channel.DataEffect(e.Effects[0].(VolumeSlide)))
}
