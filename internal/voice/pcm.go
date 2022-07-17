package voice

import (
	"time"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"
	"github.com/gotracker/voice"
	"github.com/gotracker/voice/component"
	"github.com/gotracker/voice/fadeout"
	"github.com/gotracker/voice/period"

	"github.com/gotracker/gotracker/internal/filter"
	"github.com/gotracker/gotracker/internal/pan"
	"github.com/gotracker/gotracker/internal/song/instrument"
	"github.com/gotracker/gotracker/internal/song/note"
)

// PCM is a PCM voice interface
type PCM interface {
	voice.Voice
	voice.Positioner
	voice.FreqModulator
	voice.AmpModulator
	voice.PanModulator
	voice.VolumeEnveloper
	voice.PitchEnveloper
	voice.PanEnveloper
	voice.FilterEnveloper
}

// PCMConfiguration is the information needed to configure an PCM2 voice
type PCMConfiguration struct {
	C2SPD         note.C2SPD
	InitialVolume volume.Volume
	InitialPeriod period.Period
	AutoVibrato   voice.AutoVibrato
	DataIntf      instrument.DataIntf
	OutputFilter  voice.FilterApplier
	VoiceFilter   filter.Filter
	PluginFilter  filter.Filter
}

// == the actual pcm voice ==

type pcmVoice struct {
	c2spd         note.C2SPD
	initialVolume volume.Volume
	outputFilter  voice.FilterApplier
	voiceFilter   filter.Filter
	pluginFilter  filter.Filter
	fadeoutMode   fadeout.Mode
	channels      int

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
	vol0ticks int
	done      bool
}

// NewPCM creates a new PCM voice
func NewPCM(config PCMConfiguration) voice.Voice {
	v := pcmVoice{
		c2spd:         config.C2SPD,
		initialVolume: config.InitialVolume,
		outputFilter:  config.OutputFilter,
		voiceFilter:   config.VoiceFilter,
		pluginFilter:  config.PluginFilter,
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
		v.channels = d.Sample.Channels()
	}

	v.amp.SetVolume(config.InitialVolume)
	v.freq.SetPeriod(config.InitialPeriod)
	v.freq.SetAutoVibratoEnabled(config.AutoVibrato.Enabled)
	if config.AutoVibrato.Enabled {
		v.freq.ConfigureAutoVibrato(config.AutoVibrato)
		v.freq.ResetAutoVibrato(config.AutoVibrato.Sweep)
	}

	var o PCM = &v
	return o
}

// == Controller ==

func (v *pcmVoice) Attack() {
	v.keyOn = true
	v.vol0ticks = 0
	v.done = false
	v.amp.Attack()
	v.freq.ResetAutoVibrato()
	v.sampler.Attack()
	v.SetVolumeEnvelopePosition(0)
	v.SetPitchEnvelopePosition(0)
	v.SetPanEnvelopePosition(0)
	v.SetFilterEnvelopePosition(0)
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
	if v.done {
		return true
	}

	if v.amp.IsFadeoutEnabled() {
		return v.amp.GetFadeoutVolume() <= 0
	}

	return v.vol0ticks >= 3
}

// == SampleStream ==

func (v *pcmVoice) GetSample(pos sampling.Pos) volume.Matrix {
	samp := v.sampler.GetSample(pos)
	if samp.Channels == 0 {
		v.done = true
		samp.Channels = v.channels
	}
	vol := v.GetFinalVolume()
	wet := samp.Apply(vol)
	if v.voiceFilter != nil {
		wet = v.voiceFilter.Filter(wet)
	}
	if v.pluginFilter != nil {
		wet = v.pluginFilter.Filter(wet)
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

func (v *pcmVoice) SetPeriod(period period.Period) {
	v.freq.SetPeriod(period)
}

func (v *pcmVoice) GetPeriod() period.Period {
	return v.freq.GetPeriod()
}

func (v *pcmVoice) SetPeriodDelta(delta period.Delta) {
	v.freq.SetDelta(delta)
}

func (v *pcmVoice) GetPeriodDelta() period.Delta {
	return v.freq.GetDelta()
}

func (v *pcmVoice) GetFinalPeriod() period.Period {
	p := v.freq.GetFinalPeriod()
	if v.IsPitchEnvelopeEnabled() {
		delta := v.GetCurrentPitchEnvelope()
		p = p.AddDelta(delta)
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
	if doneCB := v.volEnv.SetEnvelopePosition(pos); doneCB != nil {
		doneCB(v)
	}
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

func (v *pcmVoice) GetCurrentPitchEnvelope() period.Delta {
	if v.pitchEnv.IsEnabled() {
		return v.pitchEnv.GetCurrentValue()
	}
	return 0
}

func (v *pcmVoice) SetPitchEnvelopePosition(pos int) {
	if !v.pitchAndFilterEnvShared || !v.filterEnvActive {
		if doneCB := v.pitchEnv.SetEnvelopePosition(pos); doneCB != nil {
			doneCB(v)
		}
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

func (v *pcmVoice) GetCurrentFilterEnvelope() int8 {
	return v.filterEnv.GetCurrentValue()
}

func (v *pcmVoice) SetFilterEnvelopePosition(pos int) {
	if !v.pitchAndFilterEnvShared || v.filterEnvActive {
		if doneCB := v.filterEnv.SetEnvelopePosition(pos); doneCB != nil {
			doneCB(v)
		}
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
	if doneCB := v.panEnv.SetEnvelopePosition(pos); doneCB != nil {
		doneCB(v)
	}
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
		if doneCB := v.volEnv.Advance(v.keyOn, v.prevKeyOn); doneCB != nil {
			doneCB(v)
		}
	}
	if v.IsPanEnvelopeEnabled() {
		if doneCB := v.panEnv.Advance(v.keyOn, v.prevKeyOn); doneCB != nil {
			doneCB(v)
		}
	}
	if v.IsPitchEnvelopeEnabled() {
		if doneCB := v.pitchEnv.Advance(v.keyOn, v.prevKeyOn); doneCB != nil {
			doneCB(v)
		}
	}
	if v.IsFilterEnvelopeEnabled() {
		if doneCB := v.filterEnv.Advance(v.keyOn, v.prevKeyOn); doneCB != nil {
			doneCB(v)
		}
	}

	if v.voiceFilter != nil && v.IsFilterEnvelopeEnabled() {
		fval := v.GetCurrentFilterEnvelope()
		v.voiceFilter.UpdateEnv(fval)
	}

	if vol := v.GetFinalVolume(); vol <= 0 {
		v.vol0ticks++
	} else {
		v.vol0ticks = 0
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

func (v *pcmVoice) Clone() voice.Voice {
	p := *v
	if p.voiceFilter != nil {
		p.voiceFilter = p.voiceFilter.Clone()
	}
	if p.pluginFilter != nil {
		p.pluginFilter = p.pluginFilter.Clone()
	}
	return &p
}

func (v *pcmVoice) StartTransaction() voice.Transaction {
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
