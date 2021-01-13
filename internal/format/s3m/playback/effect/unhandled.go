package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// UnhandledCommand is an unhandled command
type UnhandledCommand struct {
	intf.Effect
	Command uint8
	Info    uint8
}

// PreStart triggers when the effect enters onto the channel state
func (e UnhandledCommand) PreStart(cs intf.Channel, p intf.Playback) {
	panic("unhandled command")
}

func (e UnhandledCommand) String() string {
	return fmt.Sprintf("%c%0.2x", e.Command+'@', e.Info)
}
