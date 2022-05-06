package effect

import (
	"fmt"

	"github.com/gotracker/gotracker/internal/format/it/layout/channel"
	effectIntf "github.com/gotracker/gotracker/internal/format/it/playback/effect/intf"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// FinePatternDelay defines an fine pattern delay effect
type FinePatternDelay channel.DataEffect // 'S6x'

// Start triggers on the first tick, but before the Tick() function is called
func (e FinePatternDelay) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()

	x := channel.DataEffect(e) & 0xf

	m := p.(effectIntf.IT)
	if err := m.AddRowTicks(int(x)); err != nil {
		return err
	}
	return nil
}

func (e FinePatternDelay) String() string {
	return fmt.Sprintf("S%0.2x", channel.DataEffect(e))
}
