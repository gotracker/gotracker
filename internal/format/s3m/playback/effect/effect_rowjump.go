package effect

import (
	"fmt"

	"github.com/gotracker/gotracker/internal/format/s3m/layout/channel"
	"github.com/gotracker/gotracker/internal/player/intf"
	"github.com/gotracker/gotracker/internal/song/index"
)

// RowJump defines a row jump effect
type RowJump ChannelCommand // 'C'

// Start triggers on the first tick, but before the Tick() function is called
func (e RowJump) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e RowJump) Stop(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback, lastTick int) error {
	r := channel.DataEffect(e)
	rowIdx := index.Row((r >> 4) * 10)
	rowIdx += index.Row(r & 0xf)
	return p.SetNextRow(rowIdx)
}

func (e RowJump) String() string {
	return fmt.Sprintf("C%0.2x", channel.DataEffect(e))
}
