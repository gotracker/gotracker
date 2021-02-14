package voice

import (
	"time"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/envelope"
	"gotracker/internal/oscillator"
	"gotracker/internal/pan"
	"gotracker/internal/player/note"
	"gotracker/internal/voice/internal/component"
)

// PCM2 is an PCM2 voice interface
type PCM2 interface {
	Voice
	FreqModulator
	AmpModulator
	PanModulator
	VolumeEnveloper
	PitchEnveloper
	PanEnveloper
	FilterEnveloper
}

// PCMConfiguration is the information needed to configure an PCM2 voice
type PCMConfiguration struct {
	C2SPD                   note.C2SPD
	VolEnv                  *envelope.Envelope
	PitchEnv                *envelope.Envelope
	PanEnv                  *envelope.Envelope
	FilterEnv               *envelope.Envelope
	PitchAndFilterEnvShared bool
	FilterEnvActive         bool // if PitchAndFilterEnvShared is true, this dictates which is active initially - true=filter, false=pitch
	FadeoutAmount           volume.Volume
	AutoVibratoSweep        int
	AutoVibrato             oscillator.Oscillator
	AutoVibratoRate         int
	AutoVibratoDepth        float32
}

// == the actual pcm voice ==

type pcm struct {
	c2spd                   note.C2SPD
	keyOn                   bool
	prevKeyOn               bool
	pitchAndFilterEnvShared bool
	filterEnvActive         bool // if pitchAndFilterEnvShared is true, this dictates which is active initially - true=filter, false=pitch
	amp                     component.AmpModulator
	freq                    component.FreqModulator
	pan                     component.PanModulator
	volEnv                  component.VolumeEnvelope
	pitchEnv                component.PitchEnvelope
	panEnv                  component.PanEnvelope
	filterEnv               component.FilterEnvelope
}

// NewPCM2 creates a new PCM2 voice
func NewPCM2(config PCMConfiguration) Voice {
	v := pcm{
		c2spd: config.C2SPD,
	}

	v.amp.ResetFadeoutValue(config.FadeoutAmount)
	v.freq.ConfigureAutoVibrato(config.AutoVibrato, config.AutoVibratoRate, config.AutoVibratoDepth)
	v.freq.ResetAutoVibrato(config.AutoVibratoSweep)
	v.volEnv.Reset(config.VolEnv)
	v.pitchEnv.Reset(config.PitchEnv)

	var o PCM2 = &v
	return o
}

// == Controller ==

func (v *pcm) Attack() {
	v.keyOn = true
	v.amp.ResetFadeoutValue()
	v.amp.SetFadeoutEnabled(false)
}

func (v *pcm) Release() {
	v.keyOn = false
}

func (v pcm) Fadeout() {
	v.amp.SetFadeoutEnabled(true)
}

func (v pcm) IsKeyOn() bool {
	return v.keyOn
}

func (v pcm) IsFadeout() bool {
	return v.amp.IsFadeoutEnabled()
}

// == FreqModulator ==

func (v *pcm) SetPeriod(period note.Period) {
	v.freq.SetPeriod(period)
}

func (v pcm) GetPeriod() note.Period {
	return v.freq.GetPeriod()
}

func (v *pcm) SetPeriodDelta(delta note.PeriodDelta) {
	v.freq.SetDelta(delta)
}

func (v pcm) GetPeriodDelta() note.PeriodDelta {
	return v.freq.GetDelta()
}

func (v pcm) GetFinalPeriod() note.Period {
	return v.freq.GetFinalPeriod().Add(v.GetCurrentPitchEnvelope())
}

// == AmpModulator ==

func (v *pcm) SetVolume(vol volume.Volume) {
	v.amp.SetVolume(vol)
}

func (v pcm) GetVolume() volume.Volume {
	return v.amp.GetVolume()
}

func (v pcm) GetFinalVolume() volume.Volume {
	return v.amp.GetFinalVolume() * v.GetCurrentVolumeEnvelope()
}

// == PanModulator ==

func (v *pcm) SetPan(pan panning.Position) {
	v.pan.SetPan(pan)
}

func (v pcm) GetPan() panning.Position {
	return v.pan.GetPan()
}

func (v pcm) GetFinalPan() panning.Position {
	return pan.CalculateCombinedPanning(v.pan.GetFinalPan(), v.panEnv.GetCurrentValue())
}

// == VolumeEnveloper ==

func (v *pcm) EnableVolumeEnvelope(enabled bool) {
	v.volEnv.SetEnabled(enabled)
}

func (v pcm) IsVolumeEnvelopeEnabled() bool {
	return v.volEnv.IsEnabled()
}

func (v pcm) GetCurrentVolumeEnvelope() volume.Volume {
	if v.volEnv.IsEnabled() {
		return v.volEnv.GetCurrentValue()
	}
	return 0
}

// == PitchEnveloper ==

func (v *pcm) EnablePitchEnvelope(enabled bool) {
	v.pitchEnv.SetEnabled(enabled)
}

func (v pcm) IsPitchEnvelopeEnabled() bool {
	if v.pitchAndFilterEnvShared && v.filterEnvActive {
		return false
	}
	return v.pitchEnv.IsEnabled()
}

func (v pcm) GetCurrentPitchEnvelope() note.PeriodDelta {
	if v.pitchEnv.IsEnabled() {
		return v.pitchEnv.GetCurrentValue()
	}
	return 0
}

// == FilterEnveloper ==

func (v *pcm) EnableFilterEnvelope(enabled bool) {
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

func (v pcm) IsFilterEnvelopeEnabled() bool {
	if v.pitchAndFilterEnvShared && !v.filterEnvActive {
		return false
	}
	return v.filterEnv.IsEnabled()
}

func (v pcm) GetCurrentFilterEnvelope() float32 {
	return v.filterEnv.GetCurrentValue()
}

// == PanEnveloper ==

func (v *pcm) EnablePanEnvelope(enabled bool) {
	v.panEnv.SetEnabled(enabled)
}

func (v pcm) IsPanEnvelopeEnabled() bool {
	return v.panEnv.IsEnabled()
}

func (v pcm) GetCurrentPanEnvelope() panning.Position {
	return v.panEnv.GetCurrentValue()
}

// == required function interfaces ==

func (v *pcm) Advance(channel int, tickDuration time.Duration) {
	defer func() {
		v.prevKeyOn = v.keyOn
	}()
	v.amp.Advance()
	v.freq.Advance()
	v.pan.Advance()
	v.volEnv.Advance(v.keyOn, v.prevKeyOn)
	v.panEnv.Advance(v.keyOn, v.prevKeyOn)
	if !v.pitchAndFilterEnvShared {
		v.pitchEnv.Advance(v.keyOn, v.prevKeyOn)
		v.filterEnv.Advance(v.keyOn, v.prevKeyOn)
	} else if v.filterEnvActive {
		v.filterEnv.Advance(v.keyOn, v.prevKeyOn)
	} else {
		v.pitchEnv.Advance(v.keyOn, v.prevKeyOn)
	}

	// determine PCM value modifications
	//vol := v.GetFinalVolume()
	//period := v.GetFinalPeriod()
	//pan := v.GetFinalPan()
}
