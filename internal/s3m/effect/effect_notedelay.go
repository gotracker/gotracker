package effect

import (
	"fmt"
	"gotracker/internal/player/intf"
)

// NoteDelay defines a note delay effect
type NoteDelay uint8 // 'SDx'

// PreStart triggers when the effect enters onto the channel state
func (e NoteDelay) PreStart(cs intf.Channel, ss intf.Song) {
}

// Start triggers on the first tick, but before the Tick() function is called
func (e NoteDelay) Start(cs intf.Channel, ss intf.Song) {
	cs.ResetRetriggerCount()

	cs.SetNotePlayTick(int(uint8(e) & 0x0F))
}

// Tick is called on every tick
func (e NoteDelay) Tick(cs intf.Channel, ss intf.Song, currentTick int) {
}

// Stop is called on the last tick of the row, but after the Tick() function is called
func (e NoteDelay) Stop(cs intf.Channel, ss intf.Song, lastTick int) {
}

func (e NoteDelay) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
