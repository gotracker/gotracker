package component

import (
	"github.com/gotracker/gomixing/volume"
)

// AmpModulator is an amplitude (volume) modulator
type AmpModulator struct {
	vol            volume.Volume
	fadeoutEnabled bool
	fadeoutVol     volume.Volume
	fadeoutAmt     volume.Volume
}

// SetVolume sets the current volume (before fadeout calculation)
func (a *AmpModulator) SetVolume(vol volume.Volume) {
	a.vol = vol
}

// GetVolume returns the current volume (before fadeout calculation)
func (a AmpModulator) GetVolume() volume.Volume {
	return a.vol
}

// SetFadeoutEnabled sets the status of the fadeout enablement flag
func (a *AmpModulator) SetFadeoutEnabled(enabled bool) {
	a.fadeoutEnabled = enabled
}

// ResetFadeoutValue resets the current fadeout value and optionally configures the amount of fadeout
func (a *AmpModulator) ResetFadeoutValue(amount ...volume.Volume) {
	a.fadeoutVol = volume.Volume(1)
	if len(amount) > 0 {
		a.fadeoutAmt = amount[0]
	}
}

// IsFadeoutEnabled returns the status of the fadeout enablement flag
func (a AmpModulator) IsFadeoutEnabled() bool {
	return a.fadeoutEnabled
}

// GetFinalVolume returns the current volume (after fadeout calculation)
func (a AmpModulator) GetFinalVolume() volume.Volume {
	if a.fadeoutEnabled {
		return a.fadeoutVol * a.vol
	}

	return a.vol
}

// Advance advances the fadeout value by 1 tick
func (a *AmpModulator) Advance() {
	if !a.fadeoutEnabled {
		return
	}

	a.fadeoutVol -= a.fadeoutAmt
	switch {
	case a.fadeoutVol < 0:
		a.fadeoutVol = 0
	case a.fadeoutVol > 1:
		a.fadeoutVol = 1
	}
}
