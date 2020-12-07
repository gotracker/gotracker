package note

import "fmt"

type Key uint8

const (
	KeyC = Key(0 + iota)
	KeyCSharp
	KeyD
	KeyDSharp
	KeyE
	KeyF
	KeyFSharp
	KeyG
	KeyGSharp
	KeyA
	KeyASharp
	KeyB
	KeyInvalid1
	KeyInvalid2
	KeyInvalid3
	KeyInvalid4
)

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

type Octave uint8

func (o Octave) String() string {
	return fmt.Sprintf("%d", o)
}

type Note uint8

const (
	EmptyNote = Note(255)
	StopNote  = Note(254)
)

func (n Note) Key() Key {
	return Key(n & 0x0F)
}

func (n Note) Octave() Octave {
	return Octave((n & 0xF0) >> 4)
}

func (n Note) IsStop() bool {
	return n == StopNote
}

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

func (n Note) Semitone() uint8 {
	key := uint8(n.Key())
	octave := uint8(n.Octave())
	return octave*12 + key
}
