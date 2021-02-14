package voice

import (
	"gotracker/internal/player/note"
)

// PitchEnveloper is a pitch envelope interface
type PitchEnveloper interface {
	EnablePitchEnvelope(enabled bool)
	IsPitchEnvelopeEnabled() bool
	GetCurrentPitchEnvelope() note.PeriodDelta
}
