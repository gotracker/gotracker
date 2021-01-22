package effect

import (
	"fmt"

	itfile "github.com/gotracker/goaudiofile/music/tracked/it"

	"gotracker/internal/format/it/playback/util"
	"gotracker/internal/player/intf"
)

// SetCoarsePanPosition defines a set coarse pan position effect
type SetCoarsePanPosition uint8 // 'S8x'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetCoarsePanPosition) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	x := uint8(e) & 0xf

	pan := itfile.PanValue(x << 4)

	cs.SetPan(util.PanningFromIt(pan))
}

func (e SetCoarsePanPosition) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
