package instrument

import (
	"time"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"
	"github.com/gotracker/opl2"

	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/render"
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
	intf.Instrument

	Modulator OPL2OperatorData
	Carrier   OPL2OperatorData

	// ModulationFeedback returns the modulation feedback
	ModulationFeedback uint8

	// AdditiveSynthesis returns true if additive synthesis is enabled
	AdditiveSynthesis bool
}

type ym3812 struct {
	chip  render.OPL2Chip
	regB0 uint8
}

// GetSample returns the sample at position `pos` in the instrument
func (inst *OPL2) GetSample(ioc intf.NoteControl, pos sampling.Pos) volume.Matrix {
	return nil
}

// GetCurrentPanning returns the panning envelope position
func (inst *OPL2) GetCurrentPanning(ioc intf.NoteControl) panning.Position {
	return panning.CenterAhead
}

// SetEnvelopePosition sets the envelope position for the note-control
func (inst *OPL2) SetEnvelopePosition(ioc intf.NoteControl, ticks int) {
}

// Initialize completes the setup of this instrument
func (inst *OPL2) Initialize(ioc intf.NoteControl) error {
	ym := ym3812{}
	ioc.SetData(&ym)

	return nil
}

// Attack sets the key on flag for the instrument
func (inst *OPL2) Attack(ioc intf.NoteControl) {
	inst.setKeyOn(ioc, true)
}

// Release clears the key on flag for the instrument
func (inst *OPL2) Release(ioc intf.NoteControl) {
	inst.setKeyOn(ioc, false)
}

func (inst *OPL2) setKeyOn(ioc intf.NoteControl, on bool) {
	ym := ioc.GetData().(*ym3812)
	ch := ym.chip
	if ch == nil {
		p := ioc.GetPlayback().(render.OPL2Intf)
		ch = p.GetOPL2Chip()
		ym.chip = ch
	}

	if ch == nil {
		panic("no ym3812 available")
	}

	index := uint32(ioc.GetOutputChannelNum())

	// write the instrument to the channel!
	if !on {
		ym.regB0 &^= 0x20 // key off
		ch.WriteReg(0xB0|index, ym.regB0)
	} else {
		ym.regB0 |= 0x20 // key on
		ch.WriteReg(0xB0|index, ym.regB0)
	}
}

// NoteCut cuts the current playback of the instrument
func (inst *OPL2) NoteCut(ioc intf.NoteControl) {
	ioc.SetVolume(0)
	inst.Release(ioc)
}

func (inst *OPL2) getReg20(o *OPL2OperatorData) uint8 {
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

func (inst *OPL2) getReg40(o *OPL2OperatorData, vol volume.Volume) uint8 {
	mVol := uint16(vol * 64)
	oVol := uint16(o.Volume)
	totalVol := uint8(oVol * mVol / 64)
	if totalVol > 63 {
		totalVol = 63
	}
	adlVol := uint8(63) - totalVol

	reg40 := uint8(0x00)
	reg40 |= (uint8(o.KeyScaleLevel) & 0x03) << 6
	reg40 |= adlVol & 0x3f
	return reg40
}

func (inst *OPL2) getReg60(o *OPL2OperatorData) uint8 {
	reg60 := uint8(0x00)
	reg60 |= (o.AttackRate & 0x0f) << 4
	reg60 |= o.DecayRate & 0x0f
	return reg60
}

func (inst *OPL2) getReg80(o *OPL2OperatorData) uint8 {
	reg80 := uint8(0x00)
	reg80 |= (15 - (o.SustainLevel & 0x0f)) << 4
	reg80 |= o.ReleaseRate & 0x0f
	return reg80
}

func (inst *OPL2) getRegC0() uint8 {
	regC0 := uint8(0x00)
	regC0 |= 0x20 | 0x10 // right and left enable [OPL3 only]
	regC0 |= uint8(inst.ModulationFeedback&0x7) << 1
	if inst.AdditiveSynthesis {
		regC0 |= 0x01
	}
	return regC0
}

func (inst *OPL2) getRegE0(o *OPL2OperatorData) uint8 {
	regE0 := uint8(0x00)
	regE0 |= uint8(o.WaveformSelection & 0x07)
	return regE0
}

// twoOperatorMelodic
var twoOperatorMelodic = [...]uint32{
	0x00, 0x01, 0x02, 0x08, 0x09, 0x0A, 0x10, 0x11, 0x12,
	0x100, 0x101, 0x102, 0x108, 0x109, 0x10A, 0x110, 0x111, 0x112,
}

func (inst *OPL2) getChannelIndex(channelIdx int) uint32 {
	return twoOperatorMelodic[channelIdx%18]
}

func freqToFnumBlock(freq float64) (uint16, uint8) {
	fnum := uint16(1023)
	block := uint8(8)

	if freq > 6208.431 {
		return 0, 0
	}

	if freq > 3104.215 {
		block = 7
	} else if freq > 1552.107 {
		block = 6
	} else if freq > 776.053 {
		block = 5
	} else if freq > 388.026 {
		block = 4
	} else if freq > 194.013 {
		block = 3
	} else if freq > 97.006 {
		block = 2
	} else if freq > 48.503 {
		block = 1
	} else {
		block = 0
	}
	fnum = uint16(freq * float64(int(1)<<(20-block)) / opl2.OPLRATE)

	return fnum, block
}

func (inst *OPL2) periodToFreqBlock(period note.Period, c2spd note.C2SPD) (uint16, uint8) {
	modFreq := period.GetFrequency()
	freq := float64(c2spd) * float64(modFreq) / 261625

	return freqToFnumBlock(freq)
}

func (inst *OPL2) freqBlockToRegA0B0(freq uint16, block uint8) (uint8, uint8) {
	regA0 := uint8(freq)
	regB0 := uint8(uint16(freq)>>8) & 0x03
	regB0 |= (block & 0x07) << 3
	return regA0, regB0
}

// GetKeyOn gets the key on flag for the instrument
func (inst *OPL2) GetKeyOn(ioc intf.NoteControl) bool {
	ym := ioc.GetData().(*ym3812)
	return (ym.regB0 & 0x20) != 0
}

// Update advances time by the amount specified by `tickDuration`
func (inst *OPL2) Update(ioc intf.NoteControl, tickDuration time.Duration) {
	ym := ioc.GetData().(*ym3812)
	ch := ym.chip
	if ch == nil {
		p := ioc.GetPlayback().(render.OPL2Intf)
		ch = p.GetOPL2Chip()
		ym.chip = ch
	}

	if ch == nil {
		panic("no ym3812 available")
	}

	index := uint32(ioc.GetOutputChannelNum())

	mod := inst.getChannelIndex(int(index))
	car := mod + 0x03

	freq, block := inst.periodToFreqBlock(ioc.GetPeriod(), ioc.GetInstrument().GetC2Spd())
	regA0, regB0 := inst.freqBlockToRegA0B0(freq, block)

	regC0 := inst.getRegC0()

	vol := ioc.GetVolume()
	modVol := vol
	if !inst.AdditiveSynthesis {
		modVol = 1.0
	}

	modReg20 := inst.getReg20(&inst.Modulator)
	modReg40 := inst.getReg40(&inst.Modulator, modVol)
	modReg60 := inst.getReg60(&inst.Modulator)
	modReg80 := inst.getReg80(&inst.Modulator)
	modRegE0 := inst.getRegE0(&inst.Modulator)

	carReg20 := inst.getReg20(&inst.Carrier)
	carReg40 := inst.getReg40(&inst.Carrier, vol)
	carReg60 := inst.getReg60(&inst.Carrier)
	carReg80 := inst.getReg80(&inst.Carrier)
	carRegE0 := inst.getRegE0(&inst.Carrier)

	ch.WriteReg(0x20|mod, modReg20)
	ch.WriteReg(0x40|mod, modReg40)
	ch.WriteReg(0x60|mod, modReg60)
	ch.WriteReg(0x80|mod, modReg80)
	ch.WriteReg(0xE0|mod, modRegE0)

	ch.WriteReg(0xA0|index, regA0)

	ch.WriteReg(0x20|car, carReg20)
	ch.WriteReg(0x40|car, carReg40)
	ch.WriteReg(0x60|car, carReg60)
	ch.WriteReg(0x80|car, carReg80)
	ch.WriteReg(0xE0|car, carRegE0)

	ch.WriteReg(0xC0|index, regC0)

	regB0 |= ym.regB0 & 0x20 // key on bit
	ym.regB0 = regB0
	ch.WriteReg(0xB0|index, regB0)
}
