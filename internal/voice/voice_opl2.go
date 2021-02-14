package voice

import (
	"time"

	"github.com/gotracker/gomixing/volume"
	opl "github.com/gotracker/opl2"

	"gotracker/internal/envelope"
	"gotracker/internal/oscillator"
	"gotracker/internal/player/note"
	"gotracker/internal/player/render"
	"gotracker/internal/voice/internal/component"
)

// OPL2 is an OPL2 voice interface
type OPL2 interface {
	Voice
	FreqModulator
	AmpModulator
	VolumeEnveloper
	PitchEnveloper
}

// OPL2Operator is a block of values specific to configuring an OPL operator (modulator or carrier)
type OPL2Operator struct {
	Reg20 uint8
	Reg40 uint8
	Reg60 uint8
	Reg80 uint8
	RegE0 uint8
}

// OPL2Registers is a set of OPL operator configurations
type OPL2Registers struct {
	Mod   OPL2Operator
	Car   OPL2Operator
	RegC0 uint8
}

// OPLConfiguration is the information needed to configure an OPL2 voice
type OPLConfiguration struct {
	Chip             render.OPL2Chip
	Registers        OPL2Registers
	C2SPD            note.C2SPD
	InitialVolume    volume.Volume
	InitialPeriod    note.Period
	VolEnv           *envelope.Envelope
	PitchEnv         *envelope.Envelope
	FadeoutAmount    volume.Volume
	AutoVibratoSweep int
	AutoVibrato      oscillator.Oscillator
	AutoVibratoRate  int
	AutoVibratoDepth float32
}

// == the actual opl2 voice ==

type opl2 struct {
	chip      render.OPL2Chip
	reg       OPL2Registers
	c2spd     note.C2SPD
	keyOn     bool
	prevKeyOn bool
	amp       component.AmpModulator
	freq      component.FreqModulator
	volEnv    component.VolumeEnvelope
	pitchEnv  component.PitchEnvelope
}

// NewOPL2 creates a new OPL2 voice
func NewOPL2(config OPLConfiguration) Voice {
	v := opl2{
		chip:  config.Chip,
		reg:   config.Registers,
		c2spd: config.C2SPD,
	}

	v.amp.SetVolume(config.InitialVolume)
	v.amp.ResetFadeoutValue(config.FadeoutAmount)
	v.freq.SetPeriod(config.InitialPeriod)
	v.freq.ConfigureAutoVibrato(config.AutoVibrato, config.AutoVibratoRate, config.AutoVibratoDepth)
	v.freq.ResetAutoVibrato(config.AutoVibratoSweep)
	v.volEnv.Reset(config.VolEnv)
	v.pitchEnv.Reset(config.PitchEnv)

	var o OPL2 = &v
	return o
}

// == Controller ==

func (v *opl2) Attack() {
	v.keyOn = true
	v.amp.ResetFadeoutValue()
	v.amp.SetFadeoutEnabled(false)
}

func (v *opl2) Release() {
	v.keyOn = false
}

func (v opl2) Fadeout() {
	v.amp.SetFadeoutEnabled(true)
}

func (v opl2) IsKeyOn() bool {
	return v.keyOn
}

func (v opl2) IsFadeout() bool {
	return v.amp.IsFadeoutEnabled()
}

// == FreqModulator ==

func (v *opl2) SetPeriod(period note.Period) {
	v.freq.SetPeriod(period)
}

func (v opl2) GetPeriod() note.Period {
	return v.freq.GetPeriod()
}

func (v *opl2) SetPeriodDelta(delta note.PeriodDelta) {
	v.freq.SetDelta(delta)
}

func (v opl2) GetPeriodDelta() note.PeriodDelta {
	return v.freq.GetDelta()
}

func (v opl2) GetFinalPeriod() note.Period {
	return v.freq.GetFinalPeriod().Add(v.GetCurrentPitchEnvelope())
}

// == AmpModulator ==

func (v *opl2) SetVolume(vol volume.Volume) {
	v.amp.SetVolume(vol)
}

func (v opl2) GetVolume() volume.Volume {
	return v.amp.GetVolume()
}

func (v opl2) GetFinalVolume() volume.Volume {
	return v.amp.GetFinalVolume() * v.GetCurrentVolumeEnvelope()
}

// == VolumeEnveloper ==

func (v *opl2) EnableVolumeEnvelope(enabled bool) {
	v.volEnv.SetEnabled(enabled)
}

func (v opl2) IsVolumeEnvelopeEnabled() bool {
	return v.volEnv.IsEnabled()
}

func (v opl2) GetCurrentVolumeEnvelope() volume.Volume {
	if v.volEnv.IsEnabled() {
		return v.volEnv.GetCurrentValue()
	}
	return 0
}

// == PitchEnveloper ==

func (v *opl2) EnablePitchEnvelope(enabled bool) {
	v.pitchEnv.SetEnabled(enabled)
}

func (v opl2) IsPitchEnvelopeEnabled() bool {
	return v.pitchEnv.IsEnabled()
}

func (v opl2) GetCurrentPitchEnvelope() note.PeriodDelta {
	if v.pitchEnv.IsEnabled() {
		return v.pitchEnv.GetCurrentValue()
	}
	return 0
}

// == required function interfaces ==

func (v *opl2) Advance(channel int, tickDuration time.Duration) {
	defer func() {
		v.prevKeyOn = v.keyOn
	}()
	v.amp.Advance()
	v.freq.Advance()
	v.volEnv.Advance(v.keyOn, v.prevKeyOn)
	v.pitchEnv.Advance(v.keyOn, v.prevKeyOn)

	// calculate the register addressing information
	index := uint32(channel)
	mod := v.getChannelIndex(channel)
	car := mod + 0x03
	ch := v.chip

	// determine register value modifications
	carVol := v.GetFinalVolume()
	modVol := volume.Volume(1)
	if (v.reg.RegC0 & 1) != 0 {
		// not additive
		modVol = carVol
	}

	var regA0, regB0 uint8
	if v.keyOn {
		period := v.GetFinalPeriod()
		freq, block := v.periodToFreqBlock(period, v.c2spd)
		regA0, regB0 = v.freqBlockToRegA0B0(freq, block)
		regB0 |= 0x20 // key on bit
	}

	// send the voice details out to the chip
	ch.WriteReg(0x20|mod, v.reg.Mod.Reg20)
	ch.WriteReg(0x40|mod, v.calc40(v.reg.Mod.Reg40, modVol))
	ch.WriteReg(0x60|mod, v.reg.Mod.Reg60)
	ch.WriteReg(0x80|mod, v.reg.Mod.Reg80)
	ch.WriteReg(0xE0|mod, v.reg.Mod.RegE0)

	ch.WriteReg(0xA0|index, regA0)

	ch.WriteReg(0x20|car, v.reg.Car.Reg20)
	ch.WriteReg(0x40|car, v.calc40(v.reg.Car.Reg40, carVol))
	ch.WriteReg(0x60|car, v.reg.Car.Reg60)
	ch.WriteReg(0x80|car, v.reg.Car.Reg80)
	ch.WriteReg(0xE0|car, v.reg.Car.RegE0)

	ch.WriteReg(0xC0|index, v.reg.RegC0)

	ch.WriteReg(0xB0|index, regB0)
}

// == support functions ==

// twoOperatorMelodic
var twoOperatorMelodic = [...]uint32{
	0x00, 0x01, 0x02, 0x08, 0x09, 0x0A, 0x10, 0x11, 0x12,
	0x100, 0x101, 0x102, 0x108, 0x109, 0x10A, 0x110, 0x111, 0x112,
}

func (v opl2) getChannelIndex(channelIdx int) uint32 {
	return twoOperatorMelodic[channelIdx%18]
}

func (v opl2) calc40(reg40 uint8, vol volume.Volume) uint8 {
	mVol := uint16(vol * 64)
	oVol := uint16(reg40 & 0x3f)
	totalVol := uint8(oVol * mVol / 64)
	if totalVol > 63 {
		totalVol = 63
	}
	adlVol := uint8(63) - totalVol

	result := reg40 &^ 0x3f
	result |= adlVol
	return result
}

func (v opl2) periodToFreqBlock(period note.Period, c2spd note.C2SPD) (uint16, uint8) {
	modFreq := period.GetFrequency()
	freq := float64(c2spd) * float64(modFreq) / 261625

	return v.freqToFnumBlock(freq)
}

func (v opl2) freqBlockToRegA0B0(freq uint16, block uint8) (uint8, uint8) {
	regA0 := uint8(freq)
	regB0 := uint8(uint16(freq)>>8) & 0x03
	regB0 |= (block & 0x07) << 3
	return regA0, regB0
}

func (v opl2) freqToFnumBlock(freq float64) (uint16, uint8) {
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
	fnum = uint16(freq * float64(int(1)<<(20-block)) / opl.OPLRATE)

	return fnum, block
}
