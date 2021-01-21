package effect

import (
	"fmt"

	itfile "github.com/gotracker/goaudiofile/music/tracked/it"

	"gotracker/internal/format/it/playback/util"
	"gotracker/internal/player/intf"
)

// SetPanPosition defines a set pan position effect
type SetPanPosition uint8 // '8xx'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetPanPosition) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	xx := itfile.PanValue(uint8(e))

	cs.SetPan(util.PanningFromIt(xx))
}

func (e SetPanPosition) String() string {
	return fmt.Sprintf("8%0.2x", uint8(e))
}
