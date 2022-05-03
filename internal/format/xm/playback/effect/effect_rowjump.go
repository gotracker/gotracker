package effect

import (
	"fmt"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
	"gotracker/internal/song/index"
)

// RowJump defines a row jump effect
type RowJump channel.DataEffect // 'D'

// Start triggers on the first tick, but before the Tick() function is called
func (e RowJump) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e RowJump) Stop(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, lastTick int) error {
	xy := channel.DataEffect(e)
	x := xy >> 4
	y := xy & 0x0f
	row := index.Row(x*10 + y)
	if err := p.BreakOrder(); err != nil {
		return err
	}
	return p.SetNextRow(row)
}

func (e RowJump) String() string {
	return fmt.Sprintf("D%0.2x", channel.DataEffect(e))
}
