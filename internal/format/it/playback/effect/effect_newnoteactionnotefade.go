package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
	"gotracker/internal/song/note"
)

// NewNoteActionNoteFade defines a NewNoteAction: Note Fade effect
type NewNoteActionNoteFade channel.DataEffect // 'S76'

// Start triggers on the first tick, but before the Tick() function is called
func (e NewNoteActionNoteFade) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.SetNewNoteAction(note.ActionFadeout)
	return nil
}

func (e NewNoteActionNoteFade) String() string {
	return fmt.Sprintf("S%0.2x", channel.DataEffect(e))
}
