package voice

import (
	"time"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/envelope"
	"gotracker/internal/loop"
	"gotracker/internal/pan"
	"gotracker/internal/pcm"
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
	Sample                  pcm.Sample
	C2SPD                   note.C2SPD
	InitialVolume           volume.Volume
	InitialPan              panning.Position
	InitialPeriod           note.Period
	InitialPos              sampling.Pos
	MixingVolume            volume.Volume
	Loop                    loop.Loop
	SustainLoop             loop.Loop
	VolEnv                  *envelope.Envelope
	PitchEnv                *envelope.Envelope
	PanEnv                  *envelope.Envelope
	FilterEnv               *envelope.Envelope
	PitchAndFilterEnvShared bool
	FilterEnvActive         bool // if PitchAndFilterEnvShared is true, this dictates which is active initially - true=filter, false=pitch
	FadeoutAmount           volume.Volume
	AutoVibrato             voiceIntf.AutoVibrato
}

// == the actual pcm voice ==

type pcmVoice struct {
	c2spd note.C2SPD

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
		c2spd:                   config.C2SPD,
		pitchAndFilterEnvShared: config.PitchAndFilterEnvShared,
		filterEnvActive:         config.FilterEnvActive,
	}

	v.sampler.Setup(config.Sample, config.Loop, config.SustainLoop)
	v.sampler.SetPos(config.InitialPos)
	v.amp.Setup(config.MixingVolume)
	v.amp.SetVolume(config.InitialVolume)
	v.amp.ResetFadeoutValue(config.FadeoutAmount)
	v.freq.SetPeriod(config.InitialPeriod)
	v.freq.ConfigureAutoVibrato(config.AutoVibrato)
	v.freq.ResetAutoVibrato(config.AutoVibrato.Sweep)
	v.pan.SetPan(config.InitialPan)
	v.volEnv.SetEnabled(config.VolEnv.Enabled)
	v.volEnv.Reset(config.VolEnv)
	v.pitchEnv.SetEnabled(config.PitchEnv.Enabled)
	v.pitchEnv.Reset(config.PitchEnv)
	v.panEnv.SetEnabled(config.PanEnv.Enabled)
	v.panEnv.Reset(config.PanEnv)
	v.filterEnv.SetEnabled(config.FilterEnv.Enabled)
	v.filterEnv.Reset(config.FilterEnv)

	var o PCM = &v
	return o
}

// == Controller ==

func (v *pcmVoice) Attack() {
	v.keyOn = true
	v.amp.ResetFadeoutValue()
	v.amp.SetFadeoutEnabled(false)
	v.freq.ResetAutoVibrato()
}

func (v *pcmVoice) Release() {
	v.keyOn = false
}

func (v *pcmVoice) Fadeout() {
	v.amp.SetFadeoutEnabled(true)
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
	return 0
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

func (v *pcmVoice) Advance(channel int, tickDuration time.Duration) {
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
}

func (v *pcmVoice) GetSampler(samplerRate float32) sampling.Sampler {
	period := v.GetFinalPeriod()
	samplerAdd := float32(period.GetSamplerAdd(float64(samplerRate)))
	return sampling.NewSampler(v, v.GetPos(), samplerAdd)
}

func (v *pcmVoice) Clone() voiceIntf.Voice {
	p := *v
	return &p
}
