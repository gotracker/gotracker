package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
	"gotracker/internal/song/note"
)

// NewNoteActionNoteContinue defines a NewNoteAction: Note Continue effect
type NewNoteActionNoteContinue channel.DataEffect // 'S74'

// Start triggers on the first tick, but before the Tick() function is called
func (e NewNoteActionNoteContinue) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.SetNewNoteAction(note.ActionContinue)
	return nil
}

func (e NewNoteActionNoteContinue) String() string {
	return fmt.Sprintf("S%0.2x", channel.DataEffect(e))
}
