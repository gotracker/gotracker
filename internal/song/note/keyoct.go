package note

import "fmt"

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
