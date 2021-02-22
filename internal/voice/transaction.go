package voice

import (
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"
	"github.com/gotracker/voice"
	"github.com/gotracker/voice/period"

	"gotracker/internal/optional"
)

type envSettings struct {
	enabled optional.Value //bool
	pos     optional.Value //int
}

type playingMode uint8

const (
	playingModeAttack = playingMode(iota)
	playingModeRelease
)

type txn struct {
	cancelled bool
	Voice     voice.Voice

	active      optional.Value //bool
	playing     optional.Value //playingMode
	fadeout     optional.Value //struct{}
	period      optional.Value //period.Period
	periodDelta optional.Value //period.Delta
	vol         optional.Value //volume.Volume
	pos         optional.Value //sampling.Pos
	pan         optional.Value //panning.Position
	volEnv      envSettings
	pitchEnv    envSettings
	panEnv      envSettings
	filterEnv   envSettings
}

func (t *txn) SetActive(active bool) {
	t.active.Set(active)
}

func (t *txn) IsPendingActive() (bool, bool) {
	return t.active.GetBool()
}

func (t *txn) IsCurrentlyActive() bool {
	return t.Voice.IsActive()
}

// Attack sets the playing mode to Attack
func (t *txn) Attack() {
	t.playing.Set(playingModeAttack)
}

// Release sets the playing mode to Release
func (t *txn) Release() {
	t.playing.Set(playingModeRelease)
}

// Fadeout activates the voice's fade-out function
func (t *txn) Fadeout() {
	t.fadeout.Set(struct{}{})
}

// SetPeriod sets the period
func (t *txn) SetPeriod(period period.Period) {
	t.period.Set(period)
}

func (t *txn) GetPendingPeriod() (period.Period, bool) {
	if p, set := t.period.GetPeriod(); set {
		if pp, ok := p.(period.Period); ok {
			return pp, set
		}
		return nil, set
	}
	return nil, false
}

func (t *txn) GetCurrentPeriod() period.Period {
	return voice.GetPeriod(t.Voice)
}

// SetPeriodDelta sets the period delta
func (t *txn) SetPeriodDelta(delta period.Delta) {
	t.periodDelta.Set(delta)
}

func (t *txn) GetPendingPeriodDelta() (period.Delta, bool) {
	return t.periodDelta.GetPeriodDelta()
}

func (t *txn) GetCurrentPeriodDelta() period.Delta {
	return voice.GetPeriodDelta(t.Voice)
}

// SetVolume sets the volume
func (t *txn) SetVolume(vol volume.Volume) {
	t.vol.Set(vol)
}

func (t *txn) GetPendingVolume() (volume.Volume, bool) {
	return t.vol.GetVolume()
}

func (t *txn) GetCurrentVolume() volume.Volume {
	return voice.GetVolume(t.Voice)
}

// SetPos sets the position
func (t *txn) SetPos(pos sampling.Pos) {
	t.pos.Set(pos)
}

func (t *txn) GetPendingPos() (sampling.Pos, bool) {
	return t.pos.GetPosition()
}

func (t *txn) GetCurrentPos() sampling.Pos {
	return voice.GetPos(t.Voice)
}

// SetPan sets the panning position
func (t *txn) SetPan(pan panning.Position) {
	t.pan.Set(pan)
}

func (t *txn) GetPendingPan() (panning.Position, bool) {
	return t.pan.GetPanning()
}

func (t *txn) GetCurrentPan() panning.Position {
	return voice.GetPan(t.Voice)
}

// SetVolumeEnvelopePosition sets the volume envelope position
func (t *txn) SetVolumeEnvelopePosition(pos int) {
	t.volEnv.pos.Set(pos)
}

// EnableVolumeEnvelope sets the volume envelope enable flag
func (t *txn) EnableVolumeEnvelope(enabled bool) {
	t.volEnv.enabled.Set(enabled)
}

func (t *txn) IsPendingVolumeEnvelopeEnabled() (bool, bool) {
	return t.volEnv.enabled.GetBool()
}

func (t *txn) IsCurrentVolumeEnvelopeEnabled() bool {
	return voice.IsVolumeEnvelopeEnabled(t.Voice)
}

// SetPitchEnvelopePosition sets the pitch envelope position
func (t *txn) SetPitchEnvelopePosition(pos int) {
	t.pitchEnv.pos.Set(pos)
}

// EnablePitchEnvelope sets the pitch envelope enable flag
func (t *txn) EnablePitchEnvelope(enabled bool) {
	t.pitchEnv.enabled.Set(enabled)
}

// SetPanEnvelopePosition sets the panning envelope position
func (t *txn) SetPanEnvelopePosition(pos int) {
	t.panEnv.pos.Set(pos)
}

// EnablePanEnvelope sets the pan envelope enable flag
func (t *txn) EnablePanEnvelope(enabled bool) {
	t.panEnv.enabled.Set(enabled)
}

// SetFilterEnvelopePosition sets the pitch envelope position
func (t *txn) SetFilterEnvelopePosition(pos int) {
	t.filterEnv.pos.Set(pos)
}

// EnableFilterEnvelope sets the filter envelope enable flag
func (t *txn) EnableFilterEnvelope(enabled bool) {
	t.filterEnv.enabled.Set(enabled)
}

// SetAllEnvelopePositions sets all the envelope positions to the same value
func (t *txn) SetAllEnvelopePositions(pos int) {
	t.volEnv.pos.Set(pos)
	t.pitchEnv.pos.Set(pos)
	t.panEnv.pos.Set(pos)
	t.filterEnv.pos.Set(pos)
}

// ======

// Cancel cancels a pending transaction
func (t *txn) Cancel() {
	t.cancelled = true
}

// Commit commits the transaction by applying pending updates
func (t *txn) Commit() {
	if t.cancelled {
		return
	}
	t.cancelled = true

	if t.Voice == nil {
		panic("voice not initialized")
	}

	if active, ok := t.active.Get(); ok {
		t.Voice.SetActive(active.(bool))
	}

	if p, ok := t.period.Get(); ok {
		voice.SetPeriod(t.Voice, p.(period.Period))
	}

	if delta, ok := t.periodDelta.Get(); ok {
		voice.SetPeriodDelta(t.Voice, delta.(period.Delta))
	}

	if vol, ok := t.vol.Get(); ok {
		voice.SetVolume(t.Voice, vol.(volume.Volume))
	}

	if pos, ok := t.pos.Get(); ok {
		voice.SetPos(t.Voice, pos.(sampling.Pos))
	}

	if pan, ok := t.pan.Get(); ok {
		voice.SetPan(t.Voice, pan.(panning.Position))
	}

	if pos, ok := t.volEnv.pos.Get(); ok {
		voice.SetVolumeEnvelopePosition(t.Voice, pos.(int))
	}

	if enabled, ok := t.volEnv.enabled.Get(); ok {
		voice.EnableVolumeEnvelope(t.Voice, enabled.(bool))
	}

	if pos, ok := t.pitchEnv.pos.Get(); ok {
		voice.SetPitchEnvelopePosition(t.Voice, pos.(int))
	}

	if enabled, ok := t.pitchEnv.enabled.Get(); ok {
		voice.EnablePitchEnvelope(t.Voice, enabled.(bool))
	}

	if pos, ok := t.panEnv.pos.Get(); ok {
		voice.SetPanEnvelopePosition(t.Voice, pos.(int))
	}

	if enabled, ok := t.panEnv.enabled.Get(); ok {
		voice.EnablePanEnvelope(t.Voice, enabled.(bool))
	}

	if pos, ok := t.filterEnv.pos.Get(); ok {
		voice.SetFilterEnvelopePosition(t.Voice, pos.(int))
	}

	if enabled, ok := t.filterEnv.enabled.Get(); ok {
		voice.EnableFilterEnvelope(t.Voice, enabled.(bool))
	}

	if mode, ok := t.playing.Get(); ok {
		switch mode.(playingMode) {
		case playingModeAttack:
			t.Voice.Attack()
		case playingModeRelease:
			t.Voice.Release()
		}
	}

	if _, ok := t.fadeout.Get(); ok {
		t.Voice.Fadeout()
	}
}

func (t *txn) GetVoice() voice.Voice {
	return t.Voice
}

func (t *txn) Clone() voice.Transaction {
	c := *t
	return &c
}
