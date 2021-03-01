package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// NewNoteActionNoteFade defines a NewNoteAction: Note Fade effect
type NewNoteActionNoteFade uint8 // 'S76'

// Start triggers on the first tick, but before the Tick() function is called
func (e NewNoteActionNoteFade) Start(cs intf.Channel, p intf.Playback) error {
	cs.SetNewNoteAction(note.ActionFadeout)
	return nil
}

func (e NewNoteActionNoteFade) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
