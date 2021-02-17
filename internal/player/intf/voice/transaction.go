package voice

import (
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/note"
)

// Transaction is an interface for updating Voice settings
type Transaction interface {
	Cancel()
	Commit()

	Attack()
	Release()
	Fadeout()
	SetPeriod(period note.Period)
	SetPeriodDelta(delta note.PeriodDelta)
	SetVolume(vol volume.Volume)
	SetPos(pos sampling.Pos)
	SetPan(pan panning.Position)
	SetVolumeEnvelopePosition(pos int)
	EnableVolumeEnvelope(enabled bool)
	SetPitchEnvelopePosition(pos int)
	EnablePitchEnvelope(enabled bool)
	SetPanEnvelopePosition(pos int)
	EnablePanEnvelope(enabled bool)
	SetFilterEnvelopePosition(pos int)
	EnableFilterEnvelope(enabled bool)
	SetAllEnvelopePositions(pos int)
}
