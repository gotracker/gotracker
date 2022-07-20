package note

import (
	s3mfile "github.com/gotracker/goaudiofile/music/tracked/s3m"
	"github.com/gotracker/gotracker/internal/song/note"
)

// NoteFromS3MNote converts an S3M file note into a player note
func NoteFromS3MNote(sn s3mfile.Note) note.Note {
	switch {
	case sn == s3mfile.EmptyNote:
		return note.EmptyNote{}
	case sn == s3mfile.StopNote:
		return note.StopOrReleaseNote{}
	default:
		k := uint8(sn.Key()) & 0x0f
		o := uint8(sn.Octave()) & 0x0f
		if k < 12 && o < 10 {
			s := note.Semitone(o*12 + k)
			return note.Normal(s)
		}
	}
	return note.InvalidNote{}
}
