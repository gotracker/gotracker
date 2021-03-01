package effect

import (
	"fmt"

	itfile "github.com/gotracker/goaudiofile/music/tracked/it"

	"gotracker/internal/format/it/playback/util"
	"gotracker/internal/player/intf"
)

// SetPanPosition defines a set pan position effect
type SetPanPosition uint8 // 'Xxx'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetPanPosition) Start(cs intf.Channel, p intf.Playback) error {
	cs.ResetRetriggerCount()

	x := uint8(e)

	pan := itfile.PanValue(x)

	cs.SetPan(util.PanningFromIt(pan))
	return nil
}

func (e SetPanPosition) String() string {
	return fmt.Sprintf("X%0.2x", uint8(e))
}
