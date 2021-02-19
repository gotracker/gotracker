package voice

import (
	"time"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/fadeout"
	"gotracker/internal/instrument"
	"gotracker/internal/pan"
	"gotracker/internal/player/intf"
	voiceIntf "gotracker/internal/player/intf/voice"
	"gotracker/internal/player/note"
	"gotracker/internal/voice/internal/component"
)

// PCM is an PCM voice interface
type PCM interface {
	voiceIntf.Voice
	voiceIntf.Positioner
	voiceIntf.FreqModulator
	voiceIntf.AmpModulator
	voiceIntf.PanModulator
	voiceIntf.VolumeEnveloper
	voiceIntf.PitchEnveloper
	voiceIntf.PanEnveloper
	voiceIntf.FilterEnveloper
}

// PCMConfiguration is the information needed to configure an PCM2 voice
type PCMConfiguration struct {
	C2SPD         note.C2SPD
	InitialVolume volume.Volume
	InitialPeriod note.Period
	AutoVibrato   voiceIntf.AutoVibrato
	DataIntf      intf.InstrumentDataIntf
	OutputFilter  voiceIntf.FilterApplier
	VoiceFilter   intf.Filter
}

// == the actual pcm voice ==

type pcmVoice struct {
	c2spd         note.C2SPD
	initialVolume volume.Volume
	outputFilter  voiceIntf.FilterApplier
	voiceFilter   intf.Filter
	fadeoutMode   fadeout.Mode

	active    bool
	keyOn     bool
	prevKeyOn bool

	pitchAndFilterEnvShared bool
	filterEnvActive         bool // if pitchAndFilterEnvShared is true, this dictates which is active initially - true=filter, false=pitch

	sampler   component.Sampler
	amp       component.AmpModulator
	freq      component.FreqModulator
	pan       component.PanModulator
	volEnv    component.VolumeEnvelope
	pitchEnv  component.PitchEnvelope
	panEnv    component.PanEnvelope
	filterEnv component.FilterEnvelope
}

// NewPCM creates a new PCM voice
func NewPCM(config PCMConfiguration) voiceIntf.Voice {
	v := pcmVoice{
		c2spd:         config.C2SPD,
		initialVolume: config.InitialVolume,
		outputFilter:  config.OutputFilter,
		voiceFilter:   config.VoiceFilter,
		active:        true,
	}

	switch d := config.DataIntf.(type) {
	case *instrument.PCM:
		v.pitchAndFilterEnvShared = true
		v.filterEnvActive = d.PitchFiltMode
		v.sampler.Setup(d.Sample, d.Loop, d.SustainLoop)
		//v.sampler.SetPos(d.InitialPos)
		v.amp.Setup(d.MixingVolume)
		v.amp.ResetFadeoutValue(d.FadeOut.Amount)
		v.pan.SetPan(d.Panning)
		v.volEnv.SetEnabled(d.VolEnv.Enabled)
		v.volEnv.Reset(&d.VolEnv)
		v.pitchEnv.SetEnabled(d.PitchFiltEnv.Enabled)
		v.pitchEnv.Reset(&d.PitchFiltEnv)
		v.panEnv.SetEnabled(d.PanEnv.Enabled)
		v.panEnv.Reset(&d.PanEnv)
		v.filterEnv.SetEnabled(d.PitchFiltEnv.Enabled)
		v.filterEnv.Reset(&d.PitchFiltEnv)
	}

	v.amp.SetVolume(config.InitialVolume)
	v.freq.SetPeriod(config.InitialPeriod)
	v.freq.ConfigureAutoVibrato(config.AutoVibrato)
	v.freq.ResetAutoVibrato(config.AutoVibrato.Sweep)

	var o PCM = &v
	return o
}

// == Controller ==

func (v *pcmVoice) Attack() {
	v.keyOn = true
	v.amp.Attack()
	v.freq.ResetAutoVibrato()
	v.sampler.Attack()
	v.volEnv.SetEnvelopePosition(0)
	v.pitchEnv.SetEnvelopePosition(0)
	v.panEnv.SetEnvelopePosition(0)
	v.filterEnv.SetEnvelopePosition(0)
}

func (v *pcmVoice) Release() {
	v.keyOn = false
	v.amp.Release()
	v.sampler.Release()
}

func (v *pcmVoice) Fadeout() {
	switch v.fadeoutMode {
	case fadeout.ModeAlwaysActive:
		v.amp.Fadeout()
	case fadeout.ModeOnlyIfVolEnvActive:
		if v.IsVolumeEnvelopeEnabled() {
			v.amp.Fadeout()
		}
	}

	v.sampler.Fadeout()
}

func (v *pcmVoice) IsKeyOn() bool {
	return v.keyOn
}

func (v *pcmVoice) IsFadeout() bool {
	return v.amp.IsFadeoutEnabled()
}

func (v *pcmVoice) IsDone() bool {
	if !v.amp.IsFadeoutEnabled() {
		return false
	}
	return v.amp.GetFadeoutVolume() <= 0
}

// == SampleStream ==

func (v *pcmVoice) GetSample(pos sampling.Pos) volume.Matrix {
	dry := v.sampler.GetSample(pos)
	vol := v.GetFinalVolume()
	wet := dry.ApplyInSitu(vol)
	if v.voiceFilter != nil {
		wet = v.voiceFilter.Filter(wet)
	}
	return wet
}

// == Positioner ==

func (v *pcmVoice) SetPos(pos sampling.Pos) {
	v.sampler.SetPos(pos)
}

func (v *pcmVoice) GetPos() sampling.Pos {
	return v.sampler.GetPos()
}

// == FreqModulator ==

func (v *pcmVoice) SetPeriod(period note.Period) {
	v.freq.SetPeriod(period)
}

func (v *pcmVoice) GetPeriod() note.Period {
	return v.freq.GetPeriod()
}

func (v *pcmVoice) SetPeriodDelta(delta note.PeriodDelta) {
	v.freq.SetDelta(delta)
}

func (v *pcmVoice) GetPeriodDelta() note.PeriodDelta {
	return v.freq.GetDelta()
}

func (v *pcmVoice) GetFinalPeriod() note.Period {
	p := v.freq.GetFinalPeriod()
	if v.IsPitchEnvelopeEnabled() {
		p = p.Add(v.GetCurrentPitchEnvelope())
	}
	return p
}

// == AmpModulator ==

func (v *pcmVoice) SetVolume(vol volume.Volume) {
	if vol == volume.VolumeUseInstVol {
		vol = v.initialVolume
	}
	v.amp.SetVolume(vol)
}

func (v *pcmVoice) GetVolume() volume.Volume {
	return v.amp.GetVolume()
}

func (v *pcmVoice) GetFinalVolume() volume.Volume {
	vol := v.amp.GetFinalVolume()
	if v.IsVolumeEnvelopeEnabled() {
		vol *= v.GetCurrentVolumeEnvelope()
	}
	return vol
}

// == PanModulator ==

func (v *pcmVoice) SetPan(pan panning.Position) {
	v.pan.SetPan(pan)
}

func (v *pcmVoice) GetPan() panning.Position {
	return v.pan.GetPan()
}

func (v *pcmVoice) GetFinalPan() panning.Position {
	p := v.pan.GetFinalPan()
	if v.IsPanEnvelopeEnabled() {
		p = pan.CalculateCombinedPanning(p, v.panEnv.GetCurrentValue())
	}
	return p
}

// == VolumeEnveloper ==

func (v *pcmVoice) EnableVolumeEnvelope(enabled bool) {
	v.volEnv.SetEnabled(enabled)
}

func (v *pcmVoice) IsVolumeEnvelopeEnabled() bool {
	return v.volEnv.IsEnabled()
}

func (v *pcmVoice) GetCurrentVolumeEnvelope() volume.Volume {
	if v.volEnv.IsEnabled() {
		return v.volEnv.GetCurrentValue()
	}
	return 1
}

func (v *pcmVoice) SetVolumeEnvelopePosition(pos int) {
	v.volEnv.SetEnvelopePosition(pos)
}

// == PitchEnveloper ==

func (v *pcmVoice) EnablePitchEnvelope(enabled bool) {
	v.pitchEnv.SetEnabled(enabled)
}

func (v *pcmVoice) IsPitchEnvelopeEnabled() bool {
	if v.pitchAndFilterEnvShared && v.filterEnvActive {
		return false
	}
	return v.pitchEnv.IsEnabled()
}

func (v *pcmVoice) GetCurrentPitchEnvelope() note.PeriodDelta {
	if v.pitchEnv.IsEnabled() {
		return v.pitchEnv.GetCurrentValue()
	}
	return 0
}

func (v *pcmVoice) SetPitchEnvelopePosition(pos int) {
	if !v.pitchAndFilterEnvShared || !v.filterEnvActive {
		v.pitchEnv.SetEnvelopePosition(pos)
	}
}

// == FilterEnveloper ==

func (v *pcmVoice) EnableFilterEnvelope(enabled bool) {
	if !v.pitchAndFilterEnvShared {
		v.filterEnv.SetEnabled(enabled)
		return
	}

	// shared filter/pitch envelope
	if !v.filterEnvActive {
		return
	}

	v.filterEnv.SetEnabled(enabled)
}

func (v *pcmVoice) IsFilterEnvelopeEnabled() bool {
	if v.pitchAndFilterEnvShared && !v.filterEnvActive {
		return false
	}
	return v.filterEnv.IsEnabled()
}

func (v *pcmVoice) GetCurrentFilterEnvelope() float32 {
	return v.filterEnv.GetCurrentValue()
}

func (v *pcmVoice) SetFilterEnvelopePosition(pos int) {
	if !v.pitchAndFilterEnvShared || v.filterEnvActive {
		v.filterEnv.SetEnvelopePosition(pos)
	}
}

// == PanEnveloper ==

func (v *pcmVoice) EnablePanEnvelope(enabled bool) {
	v.panEnv.SetEnabled(enabled)
}

func (v *pcmVoice) IsPanEnvelopeEnabled() bool {
	return v.panEnv.IsEnabled()
}

func (v *pcmVoice) GetCurrentPanEnvelope() panning.Position {
	return v.panEnv.GetCurrentValue()
}

func (v *pcmVoice) SetPanEnvelopePosition(pos int) {
	v.panEnv.SetEnvelopePosition(pos)
}

// == required function interfaces ==

func (v *pcmVoice) Advance(tickDuration time.Duration) {
	defer func() {
		v.prevKeyOn = v.keyOn
	}()
	v.amp.Advance()
	v.freq.Advance()
	v.pan.Advance()
	if v.IsVolumeEnvelopeEnabled() {
		v.volEnv.Advance(v.keyOn, v.prevKeyOn)
	}
	if v.IsPanEnvelopeEnabled() {
		v.panEnv.Advance(v.keyOn, v.prevKeyOn)
	}
	if v.IsPitchEnvelopeEnabled() {
		v.pitchEnv.Advance(v.keyOn, v.prevKeyOn)
	}
	if v.IsFilterEnvelopeEnabled() {
		v.filterEnv.Advance(v.keyOn, v.prevKeyOn)
	}

	if v.voiceFilter != nil && v.IsFilterEnvelopeEnabled() {
		fval := v.GetCurrentFilterEnvelope()
		v.voiceFilter.UpdateEnv(fval)
	}
}

func (v *pcmVoice) GetSampler(samplerRate float32) sampling.Sampler {
	period := v.GetFinalPeriod()
	samplerAdd := float32(period.GetSamplerAdd(float64(samplerRate)))
	o := component.OutputFilter{
		Input:  v,
		Output: v.outputFilter,
	}
	return sampling.NewSampler(&o, v.GetPos(), samplerAdd)
}

func (v *pcmVoice) Clone() voiceIntf.Voice {
	p := *v
	return &p
}

func (v *pcmVoice) StartTransaction() voiceIntf.Transaction {
	t := txn{
		Voice: v,
	}
	return &t
}

func (v *pcmVoice) SetActive(active bool) {
	v.active = active
}

func (v *pcmVoice) IsActive() bool {
	return v.active
}
