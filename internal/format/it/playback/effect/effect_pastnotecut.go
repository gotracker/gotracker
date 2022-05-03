package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
	"gotracker/internal/song/note"
)

// PastNoteCut defines a past note cut effect
type PastNoteCut channel.DataEffect // 'S70'

// Start triggers on the first tick, but before the Tick() function is called
func (e PastNoteCut) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.DoPastNoteEffect(note.ActionCut)
	return nil
}

func (e PastNoteCut) String() string {
	return fmt.Sprintf("S%0.2x", channel.DataEffect(e))
}
