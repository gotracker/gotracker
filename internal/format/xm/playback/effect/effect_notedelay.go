package effect

import (
	"fmt"

	"github.com/gotracker/gotracker/internal/format/xm/layout/channel"
	"github.com/gotracker/gotracker/internal/player/intf"
	"github.com/gotracker/gotracker/internal/song/note"
)

// NoteDelay defines a note delay effect
type NoteDelay channel.DataEffect // 'EDx'

// PreStart triggers when the effect enters onto the channel state
func (e NoteDelay) PreStart(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.SetNotePlayTick(true, note.ActionRetrigger, int(channel.DataEffect(e)&0x0F))
	return nil
}

// Start triggers on the first tick, but before the Tick() function is called
func (e NoteDelay) Start(cs intf.Channel[channel.Memory, channel.Data], p intf.Playback) error {
	cs.ResetRetriggerCount()
	return nil
}

func (e NoteDelay) String() string {
	return fmt.Sprintf("E%0.2x", channel.DataEffect(e))
}
