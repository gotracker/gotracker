package effect

import (
	"fmt"

	itfile "github.com/gotracker/goaudiofile/music/tracked/it"

	"gotracker/internal/format/it/playback/util"
	"gotracker/internal/player/intf"
)

// SetCoarsePanPosition defines a set pan position effect
type SetCoarsePanPosition uint8 // 'E8x'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetCoarsePanPosition) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	xy := uint8(e)
	y := xy & 0x0F

	yp := itfile.PanValue(y << 4)

	cs.SetPan(util.PanningFromIt(yp))
}

func (e SetCoarsePanPosition) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
