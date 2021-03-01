package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// PastNoteFade defines a past note fadeout effect
type PastNoteFade uint8 // 'S72'

// Start triggers on the first tick, but before the Tick() function is called
func (e PastNoteFade) Start(cs intf.Channel, p intf.Playback) error {
	cs.DoPastNoteEffect(note.ActionFadeout)
	return nil
}

func (e PastNoteFade) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
