package effect

import (
	"fmt"

	"github.com/gotracker/gotracker/internal/format/xm/layout/channel"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// FinePortaUp defines an fine portamento up effect
type FinePortaUp channel.DataEffect // 'E1x'

// Start triggers on the first tick, but before the Tick() function is called
func (e FinePortaUp) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	cs.UnfreezePlayback()

	mem := cs.GetMemory()
	xy := mem.FinePortaUp(channel.DataEffect(e))
	y := xy & 0x0F

	return doPortaUp(cs, float32(y), 4, mem.LinearFreqSlides)
}

func (e FinePortaUp) String() string {
	return fmt.Sprintf("E%0.2x", channel.DataEffect(e))
}
