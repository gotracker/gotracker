package effect

import (
	"fmt"

	"github.com/gotracker/gotracker/internal/format/it/layout/channel"
	"github.com/gotracker/gotracker/internal/player/intf"
	"github.com/gotracker/gotracker/internal/song/note"
)

// NewNoteActionNoteCut defines a NewNoteAction: Note Cut effect
type NewNoteActionNoteCut channel.DataEffect // 'S73'

// Start triggers on the first tick, but before the Tick() function is called
func (e NewNoteActionNoteCut) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.SetNewNoteAction(note.ActionCut)
	return nil
}

func (e NewNoteActionNoteCut) String() string {
	return fmt.Sprintf("S%0.2x", channel.DataEffect(e))
}
