package note

type noteSpecial int

const (
	noteSpecialEmpty = noteSpecial(iota)
	noteSpecialRelease
	noteSpecialStop
	noteSpecialNone
	noteSpecialStopOrRelease
	noteSpecialInvalid
)

// Note is a combination of key and octave
type Note struct {
	special  noteSpecial
	semitone Semitone
}

var (
	// EmptyNote denotes an empty note
	EmptyNote = Note{special: noteSpecialEmpty}
	// ReleaseNote denotes a release for the currently-playing instrument
	ReleaseNote = Note{special: noteSpecialRelease}
	// StopNote denotes a full stop for the currently-playing instrument
	StopNote = Note{special: noteSpecialStop}
	// StopOrReleaseNote denotes an S3M-style Stop note
	// NOTE: ST3 treats a "stop" note like a combination of release (note-off) and stop (note-cut)
	// For PCM, it is a stop, but for OPL2, it is a release
	StopOrReleaseNote = Note{special: noteSpecialStopOrRelease}
	// InvalidNote denotes an invalid note
	InvalidNote = Note{special: noteSpecialInvalid}
)

// NewNote returns a note from a semitone
func NewNote(s Semitone) Note {
	return Note{
		special:  noteSpecialNone,
		semitone: s,
	}
}

// Key returns the key component of the note
func (n Note) Key() Key {
	return n.semitone.Key()
}

// Octave returns the octave component of the note
func (n Note) Octave() Octave {
	return n.semitone.Octave()
}

// IsRelease returns true if the note is a release (Note-Off)
func (n Note) IsRelease(kind InstrumentKind) bool {
	if n.special == noteSpecialRelease {
		return true
	} else if kind == InstrumentKindOPL2 && n.special == noteSpecialStopOrRelease {
		return true
	}
	return false
}

// IsStop returns true if the note is a stop (Note-Cut)
func (n Note) IsStop(kind InstrumentKind) bool {
	if n.special == noteSpecialStop {
		return true
	} else if kind == InstrumentKindPCM && n.special == noteSpecialStopOrRelease {
		return true
	}
	return false
}

// IsEmpty returns true if the note is empty
func (n Note) IsEmpty() bool {
	return n.special == noteSpecialEmpty
}

// IsInvalid returns true if the note is invalid in any way
func (n Note) IsInvalid() bool {
	return n.special == noteSpecialInvalid
}

// IsSpecial returns true if the note is special in any way
func (n Note) IsSpecial() bool {
	return n.special != noteSpecialNone
}

func (n Note) String() string {
	switch n.special {
	case noteSpecialEmpty:
		return "..."
	case noteSpecialRelease:
		return "==="
	case noteSpecialStop:
		return "^^^"
	case noteSpecialNone:
		return n.Key().String() + n.Octave().String()
	case noteSpecialStopOrRelease:
		return "^^."
	case noteSpecialInvalid:
		fallthrough
	default:
		return "???"
	}
}

// Semitone returns the semitone value for the note
func (n Note) Semitone() Semitone {
	return n.semitone
}
