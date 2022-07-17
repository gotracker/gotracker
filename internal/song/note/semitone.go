package note

import "fmt"

// Semitone is a specific note in a 12-step scale of notes / octaves
type Semitone uint8

const (
	// UnchangedSemitone is a special semitone that signifies to the player that
	// the note is not remapped to another semitone value
	UnchangedSemitone = Semitone(0xFF)
)

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

func (s Semitone) String() string {
	return fmt.Sprintf("%d (%s%d)", int(s), s.Key(), s.Octave())
}
