package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// NoteDelay defines a note delay effect
type NoteDelay uint8 // 'EDx'

// PreStart triggers when the effect enters onto the channel state
func (e NoteDelay) PreStart(cs intf.Channel, p intf.Playback) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e NoteDelay) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()

	cs.SetNotePlayTick(int(uint8(e) & 0x0F))
}

// Tick is called on every tick
func (e NoteDelay) Tick(cs intf.Channel, p intf.Playback, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e NoteDelay) Stop(cs intf.Channel, p intf.Playback, lastTick int) {
}

func (e NoteDelay) String() string {
	return fmt.Sprintf("E%0.2x", uint8(e))
}
