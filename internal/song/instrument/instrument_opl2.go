package instrument

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
	FrequencyMultiplier uint8

	// KeyScaleLevel returns the key scale level
	// Attenuates the output level of the operators towards higher pitch by the given amount
	// (disabled, 1.5 dB / octave, 3 dB / octave, 6 dB / octave).
	KeyScaleLevel uint8

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
	WaveformSelection uint8
}

// OPL2 is an OPL2/Adlib instrument
type OPL2 struct {
	Modulator OPL2OperatorData
	Carrier   OPL2OperatorData

	// ModulationFeedback returns the modulation feedback
	ModulationFeedback uint8

	// AdditiveSynthesis returns true if additive synthesis is enabled
	AdditiveSynthesis bool
}

// GetReg20 calculates the Register 0x20 value
func (o *OPL2OperatorData) GetReg20() uint8 {
	reg20 := uint8(0x00)
	if o.Tremolo {
		reg20 |= 0x80
	}
	if o.Vibrato {
		reg20 |= 0x40
	}
	if o.Sustain {
		reg20 |= 0x20
	}
	if o.KeyScaleRateSelect {
		reg20 |= 0x10
	}
	reg20 |= uint8(o.FrequencyMultiplier) & 0x0f

	return reg20
}

// GetReg40 calculates the Register 0x40 value
func (o *OPL2OperatorData) GetReg40() uint8 {
	oVol := uint8(o.Volume)
	if oVol > 63 {
		oVol = 63
	}
	vol := uint8(63) - oVol

	reg40 := uint8(0x00)
	reg40 |= (uint8(o.KeyScaleLevel) & 0x03) << 6
	reg40 |= vol & 0x3f
	return reg40
}

// GetReg60 calculates the Register 0x60 value
func (o *OPL2OperatorData) GetReg60() uint8 {
	reg60 := uint8(0x00)
	reg60 |= (o.AttackRate & 0x0f) << 4
	reg60 |= o.DecayRate & 0x0f
	return reg60
}

// GetReg80 calculates the Register 0x80 value
func (o *OPL2OperatorData) GetReg80() uint8 {
	reg80 := uint8(0x00)
	reg80 |= (15 - (o.SustainLevel & 0x0f)) << 4
	reg80 |= o.ReleaseRate & 0x0f
	return reg80
}

// GetRegC0 calculates the Register 0xC0 value
func (inst *OPL2) GetRegC0() uint8 {
	regC0 := uint8(0x00)
	regC0 |= 0x20 | 0x10 // right and left enable [OPL3 only]
	regC0 |= uint8(inst.ModulationFeedback&0x7) << 1
	if inst.AdditiveSynthesis {
		regC0 |= 0x01
	}
	return regC0
}

// GetRegE0 calculates the Register 0xE0 value
func (o *OPL2OperatorData) GetRegE0() uint8 {
	regE0 := uint8(0x00)
	regE0 |= uint8(o.WaveformSelection & 0x07)
	return regE0
}
