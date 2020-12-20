package layout

import (
	s3mfile "github.com/heucuva/goaudiofile/music/tracked/s3m"
	"github.com/heucuva/gomixing/sampling"
	"github.com/heucuva/gomixing/volume"

	"gotracker/internal/player/intf"
)

// OPL2OperatorData is the operator data for an OPL2/Adlib instrument
type OPL2OperatorData struct {
	// KeyScaleRateSelect returns true if the modulator's envelope scales with keys
	// If enabled, the envelopes of higher notes are played more quickly than those of lower notes.
	KeyScaleRateSelect bool

	// Sustain returns true if the modulator's envelope sustain is enabled
	// If enabled, the volume envelope stays at the sustain stage and does not enter the
	// release stage of the envelope until a note-off event is encountered. Otherwise, it
	// directly advances from the decay stage to the release stage without waiting for a
	// note-off event.
	Sustain bool

	// Vibrato returns true if the modulator's vibrato is enabled
	// If enabled, adds a vibrato effect with a depth of 7 cents (0.07 semitones).
	// The rate of this vibrato is a static 6.4Hz.
	Vibrato bool

	// Tremolo returns true if the modulator's tremolo is enabled
	// If enabled, adds a tremolo effect with a depth of 1dB.
	// The rate of this tremolo is a static 3.7Hz.
	Tremolo bool

	// FrequencyMultiplier returns the modulator's frequency multiplier
	// Multiplies the frequency of the operator with a value between 0.5
	// (pitched one octave down) and 15.
	FrequencyMultiplier s3mfile.OPL2Multiple

	// KeyScaleLevel returns the key scale level
	// Attenuates the output level of the operators towards higher pitch by the given amount
	// (disabled, 1.5 dB / octave, 3 dB / octave, 6 dB / octave).
	KeyScaleLevel s3mfile.OPL2KSL

	// Volume returns the modulator's volume
	// The overall volume of the operator - if the modulator is in FM mode (i.e.: NOT in
	// additive synthesis mode), this will instead be the total pitch depth.
	Volume uint8

	// AttackRate returns the modulator's attack rate
	// Specifies how fast the volume envelope fades in from silence to peak volume.
	AttackRate uint8

	// DecayRate returns the modulator's decay rate
	// Specifies how fast the volume envelope reaches the sustain volume after peaking.
	DecayRate uint8

	// SustainLevel returns the modulator's sustain level
	// Specifies at which level the volume envelope is held before it is released.
	SustainLevel uint8

	// ReleaseRate returns the modulator's release rate
	// Specifies how fast the volume envelope fades out from the sustain level.
	ReleaseRate uint8

	// WaveformSelection returns the modulator's waveform selection
	WaveformSelection s3mfile.OPL2Waveform
}

// InstrumentOPL2 is an OPL2/Adlib instrument
type InstrumentOPL2 struct {
	intf.Instrument

	Modulator OPL2OperatorData
	Carrier   OPL2OperatorData

	// ModulationFeedback returns the modulation feedback
	ModulationFeedback s3mfile.OPL2Feedback

	// AdditiveSynthesis returns true if additive synthesis is enabled
	AdditiveSynthesis bool
}

// GetSample returns the sample at position `pos` in the instrument
func (inst *InstrumentOPL2) GetSample(pos sampling.Pos) volume.Matrix {
	// Notes for implementing OPL2 simulation...
	// the clock rate of the YM3812 is the same as the NTSC Colorburst frequency (3.579545454... MHz, exactly 315/88 in MHz)
	// the clock rate of the YMF262 is quadruple that rate
	// the original sampling rate was based off the NTSC Colorburst crystal rate divided by 72, giving 49715.902777777777777777777777778 Hz
	// 315000000 / 6336 Hz

	return nil
}
