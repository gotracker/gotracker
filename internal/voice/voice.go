package voice

import "time"

// Voice is a voice interface
type Voice interface {
	Controller
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
}
