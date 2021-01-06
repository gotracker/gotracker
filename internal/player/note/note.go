package note

import "fmt"

// Period is an interface that defines a sampler period
type Period interface {
	Add(Period) Period
	Compare(Period) int // <=>
	Lerp(float64, Period) Period
	GetSamplerAdd(float64) float64
}

// ComparePeriods compares two periods, taking nil into account
func ComparePeriods(lhs Period, rhs Period) int {
	if lhs == nil {
		if rhs == nil {
			return 0
		}
		return 1
	} else if rhs == nil {
		return -1
	}

	return lhs.Compare(rhs)
}

// C2SPD defines the C-2 (or in some players cases C-4) note sampling rate
type C2SPD uint16

// Semitone is a specific note in a 12-step scale of notes / octaves
type Semitone uint8

// NewSemitone creates a semitone from a key and octave
func NewSemitone(k Key, o Octave) Semitone {
	return Semitone(uint8(o)*12 + uint8(k))
}

// Key returns the key from the Semitone
func (s Semitone) Key() Key {
	return Key(s % 12)
}

// Octave returns the octave from the Semitone
func (s Semitone) Octave() Octave {
	return Octave(s / 12)
}

// Finetune is a 1/64th of a Semitone
type Finetune int16

// Key is a note key component
type Key uint8

const (
	// KeyC is C
	KeyC = Key(0 + iota)
	// KeyCSharp is C#
	KeyCSharp
	// KeyD is D
	KeyD
	// KeyDSharp is D#
	KeyDSharp
	// KeyE is E
	KeyE
	// KeyF is F
	KeyF
	// KeyFSharp is F#
	KeyFSharp
	// KeyG is G
	KeyG
	// KeyGSharp is G#
	KeyGSharp
	// KeyA is A
	KeyA
	// KeyASharp is A#
	KeyASharp
	// KeyB is B
	KeyB
	//KeyInvalid1 is invalid
	KeyInvalid1
	//KeyInvalid2 is invalid
	KeyInvalid2
	//KeyInvalid3 is invalid
	KeyInvalid3
	//KeyInvalid4 is invalid
	KeyInvalid4
)

// IsInvalid returns true if the key is invalid
func (k Key) IsInvalid() bool {
	switch k {
	case KeyInvalid1, KeyInvalid2, KeyInvalid3, KeyInvalid4:
		return true
	default:
		return false
	}
}

func (k Key) String() string {
	switch k {
	case KeyC:
		return "C-"
	case KeyCSharp:
		return "C#"
	case KeyD:
		return "D-"
	case KeyDSharp:
		return "D#"
	case KeyE:
		return "E-"
	case KeyF:
		return "F-"
	case KeyFSharp:
		return "F#"
	case KeyG:
		return "G-"
	case KeyGSharp:
		return "G#"
	case KeyA:
		return "A-"
	case KeyASharp:
		return "A#"
	case KeyB:
		return "B-"
	default:
		return "??"
	}
}

// Octave is the octave the key is in
type Octave uint8

func (o Octave) String() string {
	return fmt.Sprintf("%X", uint8(o))
}

type noteSpecial int

const (
	noteSpecialEmpty = noteSpecial(iota)
	noteSpecialStop
	noteSpecialNone
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
	// StopNote denotes a stop for the instrument
	StopNote = Note{special: noteSpecialStop}
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

// IsStop returns true if the note is a stop
func (n Note) IsStop() bool {
	return n.special == noteSpecialStop
}

// IsEmpty returns true if the note is empty
func (n Note) IsEmpty() bool {
	return n.special == noteSpecialEmpty
}

// IsInvalid returns true if the note is invalid in any way
func (n Note) IsInvalid() bool {
	return n.special == noteSpecialInvalid
}

func (n Note) String() string {
	switch n.special {
	case noteSpecialEmpty:
		return "..."
	case noteSpecialStop:
		return "^^."
	case noteSpecialNone:
		return n.Key().String() + n.Octave().String()
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
