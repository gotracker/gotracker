package effect

import (
	"fmt"

	"github.com/gotracker/gotracker/internal/format/xm/layout/channel"
	"github.com/gotracker/gotracker/internal/format/xm/playback/util"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// UnhandledCommand is an unhandled command
type UnhandledCommand struct {
	Command uint8
	Info    channel.DataEffect
}

// PreStart triggers when the effect enters onto the channel state
func (e UnhandledCommand) PreStart(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	if !p.IgnoreUnknownEffect() {
		panic("unhandled command")
	}
	return nil
}

func (e UnhandledCommand) String() string {
	switch {
	case e.Command >= 0x00 && e.Command <= 0x09:
		return fmt.Sprintf("%c%0.2x", e.Command+'0', e.Info)
	case e.Command >= 0x0A && e.Command <= 0x23:
		return fmt.Sprintf("%c%0.2x", e.Command+'A', e.Info)
	default:
		return fmt.Sprintf("?%0.2x%0.2x?", e.Command, e.Info)
	}
}

// UnhandledVolCommand is an unhandled volume command
type UnhandledVolCommand struct {
	Vol util.VolEffect
}

// PreStart triggers when the effect enters onto the channel state
func (e UnhandledVolCommand) PreStart(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	if !p.IgnoreUnknownEffect() {
		panic("unhandled command")
	}
	return nil
}

func (e UnhandledVolCommand) String() string {
	return fmt.Sprintf("v%0.2x", e.Vol)
}
