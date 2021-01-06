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
	intf.Effect
	Vol uint8
}

// PreStart triggers when the effect enters onto the channel state
func (e UnhandledVolCommand) PreStart(cs intf.Channel, p intf.Playback) {
	panic("unhandled command")
}

// Start triggers on the first tick, but before the Tick() function is called
func (e UnhandledVolCommand) Start(cs intf.Channel, p intf.Playback) {
}

// Tick is called on every tick
func (e UnhandledVolCommand) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e UnhandledVolCommand) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e UnhandledVolCommand) String() string {
	return fmt.Sprintf("v%0.2x", e.Vol)
}
