package effect

import (
	"fmt"

	itfile "github.com/gotracker/goaudiofile/music/tracked/it"
	"github.com/gotracker/gomixing/volume"

	"github.com/gotracker/gotracker/internal/format/it/layout/channel"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// SetChannelVolume defines a set channel volume effect
type SetChannelVolume channel.DataEffect // 'Mxx'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetChannelVolume) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	xx := channel.DataEffect(e)

	cv := itfile.Volume(xx)

	vol := volume.Volume(cv.Value())
	if vol > 1 {
		vol = 1
	}

	cs.SetChannelVolume(vol)
	return nil
}

func (e SetChannelVolume) String() string {
	return fmt.Sprintf("M%0.2x", channel.DataEffect(e))
}
