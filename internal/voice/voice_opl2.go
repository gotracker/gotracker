package voice

import (
	"time"

	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/fadeout"
	"gotracker/internal/instrument"
	"gotracker/internal/player/intf"
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
	Channel       int
	C2SPD         note.C2SPD
	InitialVolume volume.Volume
	InitialPeriod note.Period
	AutoVibrato   voiceIntf.AutoVibrato
	DataIntf      intf.InstrumentDataIntf
}

// == the actual opl2 voice ==

type opl2Voice struct {
	c2spd         note.C2SPD
	initialVolume volume.Volume

	keyOn     bool
	prevKeyOn bool

	fadeoutMode fadeout.Mode

	o        component.OPL2
	amp      component.AmpModulator
	freq     component.FreqModulator
	volEnv   component.VolumeEnvelope
	pitchEnv component.PitchEnvelope
}

// NewOPL2 creates a new OPL2 voice
func NewOPL2(config OPLConfiguration) voiceIntf.Voice {
	v := opl2Voice{
		c2spd:         config.C2SPD,
		initialVolume: config.InitialVolume,
		fadeoutMode:   fadeout.ModeDisabled,
	}

	var regs component.OPL2Registers

	switch d := config.DataIntf.(type) {
	case *instrument.OPL2:
		v.amp.Setup(1)
		v.amp.ResetFadeoutValue(0)
		v.volEnv.SetEnabled(false)
		v.volEnv.Reset(nil)
		v.pitchEnv.SetEnabled(false)
		v.pitchEnv.Reset(nil)
		regs.Mod.Reg20 = d.Modulator.GetReg20()
		regs.Mod.Reg40 = d.Modulator.GetReg40()
		regs.Mod.Reg60 = d.Modulator.GetReg60()
		regs.Mod.Reg80 = d.Modulator.GetReg80()
		regs.Mod.RegE0 = d.Modulator.GetRegE0()
		regs.Car.Reg20 = d.Carrier.GetReg20()
		regs.Car.Reg40 = d.Carrier.GetReg40()
		regs.Car.Reg60 = d.Carrier.GetReg60()
		regs.Car.Reg80 = d.Carrier.GetReg80()
		regs.Car.RegE0 = d.Carrier.GetRegE0()
		regs.RegC0 = d.GetRegC0()
	default:
		_ = d
	}

	v.o.Setup(config.Chip, config.Channel, regs, config.C2SPD)
	v.amp.SetVolume(config.InitialVolume)
	v.freq.SetPeriod(config.InitialPeriod)
	v.freq.ConfigureAutoVibrato(config.AutoVibrato)
	v.freq.ResetAutoVibrato(config.AutoVibrato.Sweep)

	var o OPL2 = &v
	return o
}

// == Controller ==

func (v *opl2Voice) Attack() {
	v.keyOn = true
	v.amp.Attack()
	v.freq.ResetAutoVibrato()
	v.volEnv.SetEnvelopePosition(0)
	v.pitchEnv.SetEnvelopePosition(0)

}

func (v *opl2Voice) Release() {
	v.keyOn = false
	v.amp.Release()
	v.o.Release()
}

func (v *opl2Voice) Fadeout() {
	switch v.fadeoutMode {
	case fadeout.ModeAlwaysActive:
		v.amp.Fadeout()
	case fadeout.ModeOnlyIfVolEnvActive:
		if v.IsVolumeEnvelopeEnabled() {
			v.amp.Fadeout()
		}
	}
}

func (v *opl2Voice) IsKeyOn() bool {
	return v.keyOn
}

func (v *opl2Voice) IsFadeout() bool {
	return v.amp.IsFadeoutEnabled()
}

func (v *opl2Voice) IsDone() bool {
	if !v.amp.IsFadeoutEnabled() {
		return false
	}
	return v.amp.GetFadeoutVolume() <= 0
}

// == FreqModulator ==

func (v *opl2Voice) SetPeriod(period note.Period) {
	v.freq.SetPeriod(period)
}

func (v *opl2Voice) GetPeriod() note.Period {
	return v.freq.GetPeriod()
}

func (v *opl2Voice) SetPeriodDelta(delta note.PeriodDelta) {
	v.freq.SetDelta(delta)
}

func (v *opl2Voice) GetPeriodDelta() note.PeriodDelta {
	return v.freq.GetDelta()
}

func (v *opl2Voice) GetFinalPeriod() note.Period {
	p := v.freq.GetFinalPeriod()
	if v.IsPitchEnvelopeEnabled() {
		p = p.Add(v.GetCurrentPitchEnvelope())
	}
	return p
}

// == AmpModulator ==

func (v *opl2Voice) SetVolume(vol volume.Volume) {
	if vol == volume.VolumeUseInstVol {
		vol = v.initialVolume
	}
	v.amp.SetVolume(vol)
}

func (v *opl2Voice) GetVolume() volume.Volume {
	return v.amp.GetVolume()
}

func (v *opl2Voice) GetFinalVolume() volume.Volume {
	vol := v.amp.GetFinalVolume()
	if v.IsVolumeEnvelopeEnabled() {
		vol *= v.GetCurrentVolumeEnvelope()
	}
	return vol
}

// == VolumeEnveloper ==

func (v *opl2Voice) EnableVolumeEnvelope(enabled bool) {
	v.volEnv.SetEnabled(enabled)
}

func (v *opl2Voice) IsVolumeEnvelopeEnabled() bool {
	return v.volEnv.IsEnabled()
}

func (v *opl2Voice) GetCurrentVolumeEnvelope() volume.Volume {
	if v.volEnv.IsEnabled() {
		return v.volEnv.GetCurrentValue()
	}
	return 1
}

func (v *opl2Voice) SetVolumeEnvelopePosition(pos int) {
	v.volEnv.SetEnvelopePosition(pos)
}

// == PitchEnveloper ==

func (v *opl2Voice) EnablePitchEnvelope(enabled bool) {
	v.pitchEnv.SetEnabled(enabled)
}

func (v *opl2Voice) IsPitchEnvelopeEnabled() bool {
	return v.pitchEnv.IsEnabled()
}

func (v *opl2Voice) GetCurrentPitchEnvelope() note.PeriodDelta {
	if v.pitchEnv.IsEnabled() {
		return v.pitchEnv.GetCurrentValue()
	}
	return 0
}

func (v *opl2Voice) SetPitchEnvelopePosition(pos int) {
	v.pitchEnv.SetEnvelopePosition(pos)
}

// == required function interfaces ==

func (v *opl2Voice) Advance(tickDuration time.Duration) {
	defer func() {
		v.prevKeyOn = v.keyOn
	}()
	v.amp.Advance()
	v.freq.Advance()
	if v.IsVolumeEnvelopeEnabled() {
		v.volEnv.Advance(v.keyOn, v.prevKeyOn)
	}
	if v.IsPitchEnvelopeEnabled() {
		v.pitchEnv.Advance(v.keyOn, v.prevKeyOn)
	}

	// has to be after the mod/env updates
	if v.keyOn != v.prevKeyOn {
		if v.keyOn {
			v.o.Attack()
		} else {
			v.o.Release()
		}
	}

	v.o.Advance(v.GetFinalVolume(), v.GetFinalPeriod())
}

func (v *opl2Voice) GetSample(pos sampling.Pos) volume.Matrix {
	return nil
}

func (v *opl2Voice) GetSampler(samplerRate float32, out voiceIntf.FilterApplier) sampling.Sampler {
	return nil
}

func (v *opl2Voice) Clone() voiceIntf.Voice {
	o := *v
	return &o
}
