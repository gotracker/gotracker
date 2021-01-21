package effect

import (
	"fmt"

	itfile "github.com/gotracker/goaudiofile/music/tracked/it"

	"gotracker/internal/format/it/playback/util"
	"gotracker/internal/player/intf"
)

// SetGlobalVolume defines a set global volume effect
type SetGlobalVolume uint8 // 'G'

// PreStart triggers when the effect enters onto the channel state
func (e SetGlobalVolume) PreStart(cs intf.Channel, p intf.Playback) {
	ev := itfile.Volume(e)
	p.SetGlobalVolume(util.VolumeFromIt(ev))
}

// Start triggers on the first tick, but before the Tick() function is called
func (e SetGlobalVolume) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
}

func (e SetGlobalVolume) String() string {
	return fmt.Sprintf("G%0.2x", uint8(e))
}
