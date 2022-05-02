package effect

import (
	"fmt"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/intf"
	"gotracker/internal/song/note"
)

// PastNoteOff defines a past note off effect
type PastNoteOff uint8 // 'S71'

// Start triggers on the first tick, but before the Tick() function is called
func (e PastNoteOff) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.DoPastNoteEffect(note.ActionRelease)
	return nil
}

func (e PastNoteOff) String() string {
	return fmt.Sprintf("S%0.2x", uint8(e))
}
