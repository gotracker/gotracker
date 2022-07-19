package note

import (
	"github.com/gotracker/gotracker/internal/song/note"
)

// FromXmNote converts an xm file note into a player note
func FromXmNote(in uint8) note.Note {
	switch {
	case in == 97:
		return note.ReleaseNote{}
	case in == 0:
		return note.EmptyNote{}
	case in > 97: // invalid
		return note.InvalidNote{}
	}

	an := uint8(in - 1)
	s := note.Semitone(an)
	return note.Normal(s)
}
