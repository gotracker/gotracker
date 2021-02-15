package voice

import (
	"github.com/gotracker/gomixing/panning"
)

// PanEnveloper is a pan envelope interface
type PanEnveloper interface {
	EnablePanEnvelope(enabled bool)
	IsPanEnvelopeEnabled() bool
	GetCurrentPanEnvelope() panning.Position
	SetPanEnvelopePosition(pos int)
}
