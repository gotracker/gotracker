package voice

import (
	"time"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/note"
)

// Voice is a voice interface
type Voice interface {
	Controller
	sampling.SampleStream
	// == optional control interfaces ==
	//Positioner
	//FreqModulator
	//AmpModulator
	//PanModulator
	//VolumeEnveloper
	//PanEnveloper
	//PitchEnveloper
	//FilterEnveloper

	// == required function interfaces ==
	Advance(channel int, tickDuration time.Duration)
	GetSampler(samplerRate float32, out FilterApplier) sampling.Sampler
	Clone() Voice
}

// Controller is the instrument actuation control interface
type Controller interface {
	Attack()
	Release()
	Fadeout()
	IsKeyOn() bool
	IsFadeout() bool
	IsDone() bool
}

// == Positioner ==

// SetPos sets the position within the positioner, if the interface for it exists on the voice
func SetPos(v Voice, pos sampling.Pos) {
	if p, ok := v.(Positioner); ok {
		p.SetPos(pos)
	}
}

// GetPos gets the position from the positioner, if the interface for it exists on the voice
func GetPos(v Voice) sampling.Pos {
	if p, ok := v.(Positioner); ok {
		return p.GetPos()
	}
	return sampling.Pos{}
}

// == FreqModulator ==

// SetPeriod sets the period into the frequency modulator, if the interface for it exists on the voice
func SetPeriod(v Voice, period note.Period) {
	if fm, ok := v.(FreqModulator); ok {
		fm.SetPeriod(period)
	}
}

// SetPeriodDelta sets the period delta into the frequency modulator, if the interface for it exists on the voice
func SetPeriodDelta(v Voice, delta note.PeriodDelta) {
	if fm, ok := v.(FreqModulator); ok {
		fm.SetPeriodDelta(delta)
	}
}

// GetPeriodDelta returns the period delta from the frequency modulator, if the interface for it exists on the voice
func GetPeriodDelta(v Voice) note.PeriodDelta {
	if fm, ok := v.(FreqModulator); ok {
		return fm.GetPeriodDelta()
	}
	return note.PeriodDelta(0)
}

// GetFinalPeriod returns the final period from the frequency modulator, if the interface for it exists on the voice
func GetFinalPeriod(v Voice) note.Period {
	if fm, ok := v.(FreqModulator); ok {
		return fm.GetFinalPeriod()
	}
	return nil
}

// == AmpModulator ==

// SetVolume sets the period into the amplitude modulator, if the interface for it exists on the voice
func SetVolume(v Voice, vol volume.Volume) {
	if am, ok := v.(AmpModulator); ok {
		am.SetVolume(vol)
	}
}

// GetFinalVolume returns the final volume from the amplitude modulator, if the interface for it exists on the voice
func GetFinalVolume(v Voice) volume.Volume {
	if am, ok := v.(AmpModulator); ok {
		return am.GetFinalVolume()
	}
	return volume.Volume(1)
}

// == PanModulator ==

// SetPan sets the period into the pan modulator, if the interface for it exists on the voice
func SetPan(v Voice, pan panning.Position) {
	if pm, ok := v.(PanModulator); ok {
		pm.SetPan(pan)
	}
}

// GetPan gets the period from the pan modulator, if the interface for it exists on the voice
func GetPan(v Voice) panning.Position {
	if pm, ok := v.(PanModulator); ok {
		return pm.GetPan()
	}
	return panning.CenterAhead
}

// GetFinalPan returns the final panning position from the pan modulator, if the interface for it exists on the voice
func GetFinalPan(v Voice) panning.Position {
	if pm, ok := v.(PanModulator); ok {
		return pm.GetFinalPan()
	}
	return panning.CenterAhead
}

// == VolumeEnveloper ==

// EnableVolumeEnvelope sets the volume envelope enable flag, if the interface for it exists on the voice
func EnableVolumeEnvelope(v Voice, enabled bool) {
	if ve, ok := v.(VolumeEnveloper); ok {
		ve.EnableVolumeEnvelope(enabled)
	}
}

// IsVolumeEnvelopeEnabled returns true if the volume envelope is enabled and the interface for it exists on the voice
func IsVolumeEnvelopeEnabled(v Voice) bool {
	if ve, ok := v.(VolumeEnveloper); ok {
		return ve.IsVolumeEnvelopeEnabled()
	}
	return false
}

// SetVolumeEnvelopePosition sets the volume envelope position, if the interface for it exists on the voice
func SetVolumeEnvelopePosition(v Voice, pos int) {
	if ve, ok := v.(VolumeEnveloper); ok {
		ve.SetVolumeEnvelopePosition(pos)
	}
}

// == PanEnveloper ==

// EnablePanEnvelope sets the pan envelope enable flag, if the interface for it exists on the voice
func EnablePanEnvelope(v Voice, enabled bool) {
	if pe, ok := v.(PanEnveloper); ok {
		pe.EnablePanEnvelope(enabled)
	}
}

// SetPanEnvelopePosition sets the pan envelope position, if the interface for it exists on the voice
func SetPanEnvelopePosition(v Voice, pos int) {
	if pe, ok := v.(PanEnveloper); ok {
		pe.SetPanEnvelopePosition(pos)
	}
}

// == PitchEnveloper ==

// EnablePitchEnvelope sets the pitch envelope enable flag, if the interface for it exists on the voice
func EnablePitchEnvelope(v Voice, enabled bool) {
	if pe, ok := v.(PitchEnveloper); ok {
		pe.EnablePitchEnvelope(enabled)
	}
}

// SetPitchEnvelopePosition sets the pitch envelope position, if the interface for it exists on the voice
func SetPitchEnvelopePosition(v Voice, pos int) {
	if pe, ok := v.(PitchEnveloper); ok {
		pe.SetPitchEnvelopePosition(pos)
	}
}

// == FilterEnveloper ==

// SetFilterEnvelopePosition sets the filter envelope position, if the interface for it exists on the voice
func SetFilterEnvelopePosition(v Voice, pos int) {
	if pe, ok := v.(FilterEnveloper); ok {
		pe.SetFilterEnvelopePosition(pos)
	}
}

// GetCurrentFilterEnvelope returns the filter envelope's current value, if the interface for it exists on the voice
func GetCurrentFilterEnvelope(v Voice) float32 {
	if pe, ok := v.(FilterEnveloper); ok {
		return pe.GetCurrentFilterEnvelope()
	}
	return 1
}
