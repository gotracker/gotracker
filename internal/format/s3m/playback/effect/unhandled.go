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

// Start triggers on the first tick, but before the Tick() function is called
func (e UnhandledCommand) Start(cs intf.Channel, p intf.Playback) {
}

// Tick is called on every tick
func (e UnhandledCommand) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e UnhandledCommand) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e UnhandledCommand) String() string {
	return fmt.Sprintf("%c%0.2x", e.Command+'@', e.Info)
}
