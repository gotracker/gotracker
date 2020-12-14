package note

import "fmt"

// C2SPD defines the C-2 (or in some players cases C-4) note sampling rate
type C2SPD uint16

// Semitone is a specific note in a 12-step scale of notes / octaves
type Semitone uint8

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
	return fmt.Sprintf("%d", o)
}

// Note is a combination of key and octave
type Note uint8

const (
	// EmptyNote denotes an empty note
	EmptyNote = Note(255)
	// StopNote denotes a stop for the instrument
	StopNote = Note(254)
)

// Key returns the key component of the note
func (n Note) Key() Key {
	return Key(n & 0x0F)
}

// Octave returns the octave component of the note
func (n Note) Octave() Octave {
	return Octave((n & 0xF0) >> 4)
}

// IsStop returns true if the note is a stop
func (n Note) IsStop() bool {
	return n == StopNote
}

// IsInvalid returns true if the note is invalid in any way (or is a stop)
func (n Note) IsInvalid() bool {
	return n == EmptyNote || n.IsStop() || n.Key().IsInvalid()
}

func (n Note) String() string {
	if n.IsStop() {
		return "---"
	} else if n.IsInvalid() {
		return "..."
	}
	return n.Key().String() + n.Octave().String()
}

// Semitone returns the semitone value for the note
func (n Note) Semitone() Semitone {
	key := Semitone(n.Key())
	octave := Semitone(n.Octave())
	return Semitone(octave*12 + key)
}

// Period defines a sampler period
type Period float32

// AddInteger truncates the current period to an integer and adds the delta integer in
// then returns the resulting period
func (p *Period) AddInteger(delta int) Period {
	*p = Period(int(*p) + delta)
	return *p
}

// Add adds the current period to a delta value then returns the resulting period
func (p *Period) Add(delta float32) Period {
	*p += Period(delta)
	return *p
}
