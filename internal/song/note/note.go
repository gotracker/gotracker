package note

import "fmt"

type SpecialType int

const (
	SpecialTypeEmpty = SpecialType(iota)
	SpecialTypeRelease
	SpecialTypeStop
	SpecialTypeNormal
	SpecialTypeStopOrRelease
	SpecialTypeInvalid
)

// Note is a note or special effect related to the channel's voice playback system
type Note interface {
	fmt.Stringer
	Type() SpecialType
}

type baseNote struct{}

// EmptyNote is a special note effect that specifies no change in the current voice settings
type EmptyNote baseNote

func (n EmptyNote) String() string {
	return "..."
}

// Type returns the SpecialType enumerator reflecting the type of the note
func (n EmptyNote) Type() SpecialType {
	return SpecialTypeEmpty
}

// ReleaseNote is a special note effect that releases the currently playing voice (note-off)
type ReleaseNote baseNote

func (n ReleaseNote) String() string {
	return "==="
}

// Type returns the SpecialType enumerator reflecting the type of the note
func (n ReleaseNote) Type() SpecialType {
	return SpecialTypeRelease
}

// StopNote is a special note effect that stops the currently playing voice (note-cut)
type StopNote baseNote

func (n StopNote) String() string {
	return "^^^"
}

// Type returns the SpecialType enumerator reflecting the type of the note
func (n StopNote) Type() SpecialType {
	return SpecialTypeStop
}

// Normal is a standard note, which is a combination of key and octave
type Normal Semitone

func (n Normal) String() string {
	st := Semitone(n)
	return st.Key().String() + st.Octave().String()
}

// Type returns the SpecialType enumerator reflecting the type of the note
func (n Normal) Type() SpecialType {
	return SpecialTypeNormal
}

// StopOrReleaseNote is a special note effect that denotes an S3M-style Stop note
// NOTE: ST3 treats a "stop" note like a combination of release (note-off) and stop (note-cut)
// For PCM, it is a stop, but for OPL2, it is a release
type StopOrReleaseNote baseNote

func (n StopOrReleaseNote) String() string {
	return "^^."
}

// Type returns the SpecialType enumerator reflecting the type of the note
func (n StopOrReleaseNote) Type() SpecialType {
	return SpecialTypeStopOrRelease
}

// InvalidNote is a special note effect that stops the currently playing voice (note-cut)
type InvalidNote baseNote

func (n InvalidNote) String() string {
	return "???"
}

// Type returns the SpecialType enumerator reflecting the type of the note
func (n InvalidNote) Type() SpecialType {
	return SpecialTypeInvalid
}

// CoalesceNoteSemitone will coalesce a note and an included semitone value.
// The intention is that a special note (note-off, fade-out, etc.) will take precedence
// over the semitone passed in, but if the semitone asks to override a normal note's
// semitone value, it will.
func CoalesceNoteSemitone(n Note, s Semitone) Note {
	if s == UnchangedSemitone || IsSpecial(n) {
		return n
	}

	return Normal(s)
}

// IsRelease returns true if the note is a release (Note-Off)
func IsRelease(n Note) bool {
	return n != nil && n.Type() == SpecialTypeRelease
}

// IsStop returns true if the note is a stop (Note-Cut)
func IsStop(n Note) bool {
	return n != nil && n.Type() == SpecialTypeStop
}

// IsEmpty returns true if the note is empty
func IsEmpty(n Note) bool {
	return n == nil || n.Type() == SpecialTypeEmpty
}

// IsInvalid returns true if the note is invalid in any way
func IsInvalid(n Note) bool {
	return n != nil && n.Type() == SpecialTypeInvalid
}

// IsSpecial returns true if the note is special in any way
func IsSpecial(n Note) bool {
	return n != nil && n.Type() != SpecialTypeNormal
}

// Type returns the SpecialType enumerator reflecting the type of the note
func Type(n Note) SpecialType {
	if n == nil {
		return SpecialTypeEmpty
	}

	return n.Type()
}

// String returns the string representation of the note presented
func String(n Note) string {
	if n == nil {
		return EmptyNote{}.String()
	}

	return n.String()
}
