package component

import (
	"github.com/gotracker/gomixing/volume"
)

// AmpModulator is an amplitude (volume) modulator
type AmpModulator struct {
	vol            volume.Volume
	mixing         volume.Volume
	fadeoutEnabled bool
	fadeoutVol     volume.Volume
	fadeoutAmt     volume.Volume
	final          volume.Volume // = [fadeoutVol *] mixing * vol
}

// Setup configures the initial settings of the modulator
func (a *AmpModulator) Setup(mixing volume.Volume) {
	a.mixing = mixing
	a.updateFinal()
}

// Attack disables the fadeout and resets its volume
func (a *AmpModulator) Attack() {
	a.fadeoutEnabled = false
	a.fadeoutVol = volume.Volume(1)
	a.updateFinal()
}

// Release currently does nothing
func (a *AmpModulator) Release() {
}

// Fadeout activates the fadeout
func (a *AmpModulator) Fadeout() {
	a.fadeoutEnabled = true
	a.updateFinal()
}

// SetVolume sets the current volume (before fadeout calculation)
func (a *AmpModulator) SetVolume(vol volume.Volume) {
	a.vol = vol
	a.updateFinal()
}

// GetVolume returns the current volume (before fadeout calculation)
func (a *AmpModulator) GetVolume() volume.Volume {
	return a.vol
}

// SetFadeoutEnabled sets the status of the fadeout enablement flag
func (a *AmpModulator) SetFadeoutEnabled(enabled bool) {
	a.fadeoutEnabled = enabled
	a.updateFinal()
}

// ResetFadeoutValue resets the current fadeout value and optionally configures the amount of fadeout
func (a *AmpModulator) ResetFadeoutValue(amount ...volume.Volume) {
	a.fadeoutVol = volume.Volume(1)
	if len(amount) > 0 {
		a.fadeoutAmt = amount[0]
	}
	a.updateFinal()
}

// IsFadeoutEnabled returns the status of the fadeout enablement flag
func (a *AmpModulator) IsFadeoutEnabled() bool {
	return a.fadeoutEnabled
}

// GetFadeoutVolume returns the value of the fadeout volume
func (a *AmpModulator) GetFadeoutVolume() volume.Volume {
	return a.fadeoutVol
}

// GetFinalVolume returns the current volume (after fadeout calculation)
func (a *AmpModulator) GetFinalVolume() volume.Volume {
	return a.final
}

// Advance advances the fadeout value by 1 tick
func (a *AmpModulator) Advance() {
	if a.fadeoutEnabled || a.fadeoutVol <= 0 {
		return
	}

	a.fadeoutVol -= a.fadeoutAmt
	switch {
	case a.fadeoutVol < 0:
		a.fadeoutVol = 0
	case a.fadeoutVol > 1:
		a.fadeoutVol = 1
	}
	a.updateFinal()
}

func (a *AmpModulator) updateFinal() {
	a.final = a.mixing * a.vol
	if a.fadeoutEnabled {
		a.final *= a.fadeoutVol
	}
}
