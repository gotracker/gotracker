package effect

import (
	"fmt"

	itfile "github.com/gotracker/goaudiofile/music/tracked/it"

	"github.com/gotracker/gotracker/internal/format/it/layout/channel"
	"github.com/gotracker/gotracker/internal/format/it/playback/util"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// SetCoarsePanPosition defines a set coarse pan position effect
type SetCoarsePanPosition channel.DataEffect // 'S8x'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetCoarsePanPosition) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	x := channel.DataEffect(e) & 0xf

	pan := itfile.PanValue(x << 2)

	cs.SetPan(util.PanningFromIt(pan))
	return nil
}

func (e SetCoarsePanPosition) String() string {
	return fmt.Sprintf("S%0.2x", channel.DataEffect(e))
}
