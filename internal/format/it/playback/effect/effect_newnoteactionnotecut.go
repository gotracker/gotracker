package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// NewNoteActionNoteCut defines a NewNoteAction: Note Cut effect
type NewNoteActionNoteCut uint8 // 'S73'

// Start triggers on the first tick, but before the Tick() function is called
func (e NewNoteActionNoteCut) Start(cs intf.Channel, p intf.Playback) {
	cs.SetNewNoteAction(note.ActionCut)
}

func (e NewNoteActionNoteCut) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
