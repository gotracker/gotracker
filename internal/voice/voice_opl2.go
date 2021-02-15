package voice

import (
	"time"

	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/envelope"
	voiceIntf "gotracker/internal/player/intf/voice"
	"gotracker/internal/player/note"
	"gotracker/internal/player/render"
	"gotracker/internal/voice/internal/component"
)

// OPL2 is an OPL2 voice interface
type OPL2 interface {
	voiceIntf.Voice
	voiceIntf.FreqModulator
	voiceIntf.AmpModulator
	voiceIntf.VolumeEnveloper
	voiceIntf.PitchEnveloper
}

// OPL2Registers is a set of OPL operator configurations
type OPL2Registers component.OPL2Registers

// OPLConfiguration is the information needed to configure an OPL2 voice
type OPLConfiguration struct {
	Chip          render.OPL2Chip
	Registers     OPL2Registers
	C2SPD         note.C2SPD
	InitialVolume volume.Volume
	InitialPeriod note.Period
	VolEnv        *envelope.Envelope
	PitchEnv      *envelope.Envelope
	FadeoutAmount volume.Volume
	AutoVibrato   voiceIntf.AutoVibrato
}

// == the actual opl2 voice ==

type opl2Voice struct {
	keyOn     bool
	prevKeyOn bool

	o        component.OPL2
	amp      component.AmpModulator
	freq     component.FreqModulator
	volEnv   component.VolumeEnvelope
	pitchEnv component.PitchEnvelope
}

// NewOPL2 creates a new OPL2 voice
func NewOPL2(config OPLConfiguration) voiceIntf.Voice {
	v := opl2Voice{}

	v.o.Setup(config.Chip, component.OPL2Registers(config.Registers), config.C2SPD)
	v.amp.SetVolume(config.InitialVolume)
	v.amp.ResetFadeoutValue(config.FadeoutAmount)
	v.freq.SetPeriod(config.InitialPeriod)
	v.freq.ConfigureAutoVibrato(config.AutoVibrato)
	v.freq.ResetAutoVibrato(config.AutoVibrato.Sweep)
	v.volEnv.Reset(config.VolEnv)
	v.pitchEnv.Reset(config.PitchEnv)

	var o OPL2 = &v
	return o
}

// == Controller ==

func (v *opl2Voice) Attack() {
	v.keyOn = true
	v.amp.ResetFadeoutValue()
	v.amp.SetFadeoutEnabled(false)
	v.o.Attack()
}

func (v *opl2Voice) Release() {
	v.keyOn = false
	v.o.Release()
}

func (v opl2Voice) Fadeout() {
	v.amp.SetFadeoutEnabled(true)
}

func (v opl2Voice) IsKeyOn() bool {
	return v.keyOn
}

func (v opl2Voice) IsFadeout() bool {
	return v.amp.IsFadeoutEnabled()
}

func (v opl2Voice) IsDone() bool {
	if !v.amp.IsFadeoutEnabled() {
		return false
	}
	return v.amp.GetFadeoutVolume() <= 0
}

// == FreqModulator ==

func (v *opl2Voice) SetPeriod(period note.Period) {
	v.freq.SetPeriod(period)
}

func (v opl2Voice) GetPeriod() note.Period {
	return v.freq.GetPeriod()
}

func (v *opl2Voice) SetPeriodDelta(delta note.PeriodDelta) {
	v.freq.SetDelta(delta)
}

func (v opl2Voice) GetPeriodDelta() note.PeriodDelta {
	return v.freq.GetDelta()
}

func (v opl2Voice) GetFinalPeriod() note.Period {
	return v.freq.GetFinalPeriod().Add(v.GetCurrentPitchEnvelope())
}

// == AmpModulator ==

func (v *opl2Voice) SetVolume(vol volume.Volume) {
	v.amp.SetVolume(vol)
}

func (v opl2Voice) GetVolume() volume.Volume {
	return v.amp.GetVolume()
}

func (v opl2Voice) GetFinalVolume() volume.Volume {
	return v.amp.GetFinalVolume() * v.GetCurrentVolumeEnvelope()
}

// == VolumeEnveloper ==

func (v *opl2Voice) EnableVolumeEnvelope(enabled bool) {
	v.volEnv.SetEnabled(enabled)
}

func (v opl2Voice) IsVolumeEnvelopeEnabled() bool {
	return v.volEnv.IsEnabled()
}

func (v opl2Voice) GetCurrentVolumeEnvelope() volume.Volume {
	if v.volEnv.IsEnabled() {
		return v.volEnv.GetCurrentValue()
	}
	return 0
}

func (v *opl2Voice) SetVolumeEnvelopePosition(pos int) {
	v.volEnv.SetEnvelopePosition(pos)
}

// == PitchEnveloper ==

func (v *opl2Voice) EnablePitchEnvelope(enabled bool) {
	v.pitchEnv.SetEnabled(enabled)
}

func (v opl2Voice) IsPitchEnvelopeEnabled() bool {
	return v.pitchEnv.IsEnabled()
}

func (v opl2Voice) GetCurrentPitchEnvelope() note.PeriodDelta {
	if v.pitchEnv.IsEnabled() {
		return v.pitchEnv.GetCurrentValue()
	}
	return 0
}

func (v *opl2Voice) SetPitchEnvelopePosition(pos int) {
	v.pitchEnv.SetEnvelopePosition(pos)
}

// == required function interfaces ==

func (v *opl2Voice) Advance(channel int, tickDuration time.Duration) {
	defer func() {
		v.prevKeyOn = v.keyOn
	}()
	v.amp.Advance()
	v.freq.Advance()
	v.volEnv.Advance(v.keyOn, v.prevKeyOn)
	v.pitchEnv.Advance(v.keyOn, v.prevKeyOn)

	// has to be after the mod/env updates
	v.o.Advance(channel, v.GetFinalVolume(), v.GetFinalPeriod())
}

func (v *opl2Voice) GetSample(pos sampling.Pos) volume.Matrix {
	return nil
}

func (v *opl2Voice) GetSampler(samplerRate float32) sampling.Sampler {
	return nil
}

func (v opl2Voice) Clone() voiceIntf.Voice {
	o := v
	return &o
}
