package note

type SpecialType int

const (
	SpecialTypeEmpty = SpecialType(iota)
	SpecialTypeRelease
	SpecialTypeStop
	SpecialTypeNormal
	SpecialTypeStopOrRelease
	SpecialTypeInvalid
)

// Note is a combination of key and octave
type Note struct {
	special  SpecialType
	semitone Semitone
}

var (
	// EmptyNote denotes an empty note
	EmptyNote = Note{special: SpecialTypeEmpty}
	// ReleaseNote denotes a release for the currently-playing instrument
	ReleaseNote = Note{special: SpecialTypeRelease}
	// StopNote denotes a full stop for the currently-playing instrument
	StopNote = Note{special: SpecialTypeStop}
	// StopOrReleaseNote denotes an S3M-style Stop note
	// NOTE: ST3 treats a "stop" note like a combination of release (note-off) and stop (note-cut)
	// For PCM, it is a stop, but for OPL2, it is a release
	StopOrReleaseNote = Note{special: SpecialTypeStopOrRelease}
	// InvalidNote denotes an invalid note
	InvalidNote = Note{special: SpecialTypeInvalid}
)

// NewNote returns a note from a semitone
func NewNote(s Semitone) Note {
	return Note{
		special:  SpecialTypeNormal,
		semitone: s,
	}
}

// CoalesceNoteSemitone will coalesce a note and an included semitone value
// the intention is that a special note (note-off, fade-out, etc.) will take precedence
// over the semitone passed in, but if the semitone asks to override a normal note's
// semitone value, it will.
func CoalesceNoteSemitone(n Note, s Semitone) Note {
	if s == UnchangedSemitone || n.IsSpecial() {
		return n
	}

	return NewNote(s)
}

// Key returns the key component of the note
func (n Note) Key() Key {
	return n.semitone.Key()
}

// Octave returns the octave component of the note
func (n Note) Octave() Octave {
	return n.semitone.Octave()
}

// Type returns the SpecialType enumerator reflecting the type of the note
func (n Note) Type() SpecialType {
	return n.special
}

// IsRelease returns true if the note is a release (Note-Off)
func (n Note) IsRelease() bool {
	return n.special == SpecialTypeRelease
}

// IsStop returns true if the note is a stop (Note-Cut)
func (n Note) IsStop() bool {
	return n.special == SpecialTypeStop
}

// IsEmpty returns true if the note is empty
func (n Note) IsEmpty() bool {
	return n.special == SpecialTypeEmpty
}

// IsInvalid returns true if the note is invalid in any way
func (n Note) IsInvalid() bool {
	return n.special == SpecialTypeInvalid
}

// IsSpecial returns true if the note is special in any way
func (n Note) IsSpecial() bool {
	return n.special != SpecialTypeNormal
}

func (n Note) String() string {
	switch n.special {
	case SpecialTypeEmpty:
		return "..."
	case SpecialTypeRelease:
		return "==="
	case SpecialTypeStop:
		return "^^^"
	case SpecialTypeNormal:
		return n.Key().String() + n.Octave().String()
	case SpecialTypeStopOrRelease:
		return "^^."
	case SpecialTypeInvalid:
		fallthrough
	default:
		return "???"
	}
}

// Semitone returns the semitone value for the note
func (n Note) Semitone() Semitone {
	return n.semitone
}
