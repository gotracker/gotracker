package note

// NewNoteAction is the action to take when a new note enters a channel
type NewNoteAction uint8

const (
	// NewNoteActionNoteCut will cut the old note/instrument playback immediately
	// (this is the default for most tracker formats)
	NewNoteActionNoteCut = NewNoteAction(iota)
	// NewNoteActionContinue will continue the old note/instrument playback indefinitely
	NewNoteActionContinue
	// NewNoteActionNoteOff will perform a release (key-off) on the old note/instrument playback
	NewNoteActionNoteOff
	// NewNoteActionFadeout will fade out the old note/instrument playback
	// (if the instrument's fadeout volume is 0, then this effectively becomes a NewNoteActionContinue)
	NewNoteActionFadeout
)
