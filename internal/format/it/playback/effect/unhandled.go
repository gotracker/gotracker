package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// UnhandledCommand is an unhandled command
type UnhandledCommand struct {
	Command uint8
	Info    uint8
}

// PreStart triggers when the effect enters onto the channel state
func (e UnhandledCommand) PreStart(cs intf.Channel, p intf.Playback) error {
	if !p.IgnoreUnknownEffect() {
		panic(fmt.Sprintf("unhandled command: ce:%0.2X cp:%0.2X", e.Command, e.Info))
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
	Vol uint8
}

// PreStart triggers when the effect enters onto the channel state
func (e UnhandledVolCommand) PreStart(cs intf.Channel, p intf.Playback) error {
	if !p.IgnoreUnknownEffect() {
		panic(fmt.Sprintf("unhandled command: volCmd:%0.2X", e.Vol))
	}
	return nil
}

func (e UnhandledVolCommand) String() string {
	return fmt.Sprintf("v%0.2x", e.Vol)
}
