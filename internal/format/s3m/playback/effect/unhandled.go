package effect

import (
	"fmt"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// UnhandledCommand is an unhandled command
type UnhandledCommand struct {
	Command uint8
	Info    uint8
}

// PreStart triggers when the effect enters onto the channel state
func (e UnhandledCommand) PreStart(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	if !p.IgnoreUnknownEffect() {
		panic("unhandled command")
	}
	return nil
}

func (e UnhandledCommand) String() string {
	return fmt.Sprintf("%c%0.2x", e.Command+'@', e.Info)
}
