package effect

import (
	"fmt"

	itfile "github.com/gotracker/goaudiofile/music/tracked/it"

	"gotracker/internal/format/it/playback/util"
	"gotracker/internal/player/intf"
)

// SetVolume defines a volume slide effect
type SetVolume uint8 // 'C'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetVolume) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	xx := itfile.Volume(uint8(e))

	cs.SetActiveVolume(util.VolumeFromIt(xx))
}

func (e SetVolume) String() string {
	return fmt.Sprintf("C%0.2x", uint8(e))
}
