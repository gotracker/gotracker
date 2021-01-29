package note

// Action is the action to take on a note
type Action uint8

const (
	// ActionNoteCut will cut the old note/instrument playback immediately
	// (this is the default for most tracker formats)
	ActionNoteCut = Action(iota)
	// ActionContinue will continue the old note/instrument playback indefinitely
	ActionContinue
	// ActionNoteOff will perform a release (key-off) on the old note/instrument playback
	ActionNoteOff
	// ActionFadeout will fade out the old note/instrument playback
	// (if the instrument's fadeout volume is 0, then this effectively becomes a NewNoteActionContinue)
	ActionFadeout
)
