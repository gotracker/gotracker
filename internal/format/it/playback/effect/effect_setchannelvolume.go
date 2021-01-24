package effect

import (
	"fmt"

	itfile "github.com/gotracker/goaudiofile/music/tracked/it"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/intf"
)

// SetChannelVolume defines a set channel volume effect
type SetChannelVolume uint8 // 'Mxx'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetChannelVolume) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	xx := uint8(e)

	cv := itfile.Volume(xx)

	vol := volume.Volume(cv.Value())
	if vol > 1 {
		vol = 1
	}

	cs.SetChannelVolume(vol)
}

func (e SetChannelVolume) String() string {
	return fmt.Sprintf("M%0.2x", uint8(e))
}
