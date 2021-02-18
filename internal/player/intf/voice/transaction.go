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
	GetVoice() Voice
	Clone() Transaction

	SetActive(active bool)
	IsPendingActive() (bool, bool)
	IsCurrentlyActive() bool

	Attack()
	Release()
	Fadeout()
	SetPeriod(period note.Period)
	GetPendingPeriod() (note.Period, bool)
	GetCurrentPeriod() note.Period
	SetPeriodDelta(delta note.PeriodDelta)
	GetPendingPeriodDelta() (note.PeriodDelta, bool)
	GetCurrentPeriodDelta() note.PeriodDelta
	SetVolume(vol volume.Volume)
	GetPendingVolume() (volume.Volume, bool)
	GetCurrentVolume() volume.Volume
	SetPos(pos sampling.Pos)
	GetPendingPos() (sampling.Pos, bool)
	GetCurrentPos() sampling.Pos
	SetPan(pan panning.Position)
	GetPendingPan() (panning.Position, bool)
	GetCurrentPan() panning.Position
	SetVolumeEnvelopePosition(pos int)
	EnableVolumeEnvelope(enabled bool)
	IsPendingVolumeEnvelopeEnabled() (bool, bool)
	IsCurrentVolumeEnvelopeEnabled() bool
	SetPitchEnvelopePosition(pos int)
	EnablePitchEnvelope(enabled bool)
	SetPanEnvelopePosition(pos int)
	EnablePanEnvelope(enabled bool)
	SetFilterEnvelopePosition(pos int)
	EnableFilterEnvelope(enabled bool)
	SetAllEnvelopePositions(pos int)
}
