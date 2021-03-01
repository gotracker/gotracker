package effect

import (
	"fmt"

	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// PastNoteCut defines a past note cut effect
type PastNoteCut uint8 // 'S70'

// Start triggers on the first tick, but before the Tick() function is called
func (e PastNoteCut) Start(cs intf.Channel, p intf.Playback) error {
	cs.DoPastNoteEffect(note.ActionCut)
	return nil
}

func (e PastNoteCut) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
