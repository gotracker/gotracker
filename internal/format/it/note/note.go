package note

import (
	itfile "github.com/gotracker/goaudiofile/music/tracked/it"

	"github.com/gotracker/gotracker/internal/song/note"
)

// FromItNote converts an it file note into a player note
func FromItNote(in itfile.Note) note.Note {
	switch {
	case in.IsNoteOff():
		return note.ReleaseNote{}
	case in.IsNoteCut():
		return note.StopNote{}
	case in.IsNoteFade(): // not really invalid, but...
		return note.InvalidNote{}
	}

	an := uint8(in)
	s := note.Semitone(an)
	return note.Normal(s)
}
