package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// NewNoteActionNoteOff defines a NewNoteAction: Note Off effect
type NewNoteActionNoteOff uint8 // 'S75'

// Start triggers on the first tick, but before the Tick() function is called
func (e NewNoteActionNoteOff) Start(cs intf.Channel, p intf.Playback) {
	cs.SetNewNoteAction(note.ActionRelease)
}

func (e NewNoteActionNoteOff) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
