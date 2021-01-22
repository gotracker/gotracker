package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
)

// NoteDelay defines a note delay effect
type NoteDelay uint8 // 'SDx'

// PreStart triggers when the effect enters onto the channel state
func (e NoteDelay) PreStart(cs intf.Channel, p intf.Playback) {
	cs.SetNotePlayTick(int(uint8(e) & 0x0F))
}

// Start triggers on the first tick, but before the Tick() function is called
func (e NoteDelay) Start(cs intf.Channel, p intf.Playback) {
	cs.ResetRetriggerCount()
}

func (e NoteDelay) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
