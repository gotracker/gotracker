package effect

import (
	"fmt"

	itfile "github.com/gotracker/goaudiofile/music/tracked/it"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/format/it/playback/util"
	"gotracker/internal/player/intf"
)

// SetPanPosition defines a set pan position effect
type SetPanPosition channel.DataEffect // 'Xxx'

// Start triggers on the first tick, but before the Tick() function is called
func (e SetPanPosition) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	x := channel.DataEffect(e)

	pan := itfile.PanValue(x)

	cs.SetPan(util.PanningFromIt(pan))
	return nil
}

func (e SetPanPosition) String() string {
	return fmt.Sprintf("X%0.2x", channel.DataEffect(e))
}
