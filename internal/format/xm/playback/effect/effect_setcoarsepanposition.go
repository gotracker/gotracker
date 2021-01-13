package effect

import (
	"fmt"

	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/player/intf"
)

// SetCoarsePanPosition defines a set pan position effect
type SetCoarsePanPosition uint8 // 'E8x'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetCoarsePanPosition) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	xy := uint8(e)
	y := xy & 0x0F

	cs.SetPan(util.PanningFromXm(y << 4))
}

func (e SetCoarsePanPosition) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
