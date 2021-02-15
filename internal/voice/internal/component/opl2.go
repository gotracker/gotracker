package component

import (
	"gotracker/internal/player/note"
	"gotracker/internal/player/render"

	"github.com/gotracker/gomixing/volume"
	"github.com/gotracker/opl2"
)

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

// OPL2 is an OPL2 component
type OPL2 struct {
	chip  render.OPL2Chip
	reg   OPL2Registers
	c2spd note.C2SPD
	keyOn bool
}

// Setup sets up the OPL2 component
func (o *OPL2) Setup(chip render.OPL2Chip, reg OPL2Registers, c2spd note.C2SPD) {
	o.chip = chip
	o.reg = reg
	o.c2spd = c2spd
	o.keyOn = false
}

// Attack activates the key-on bit
func (o *OPL2) Attack() {
	o.keyOn = true
}

// Release deactivates the key-on bit
func (o *OPL2) Release() {
	o.keyOn = false
}

// Advance advances the playback
func (o *OPL2) Advance(channel int, carVol volume.Volume, period note.Period) {
	// calculate the register addressing information
	index := uint32(channel)
	mod := o.getChannelIndex(channel)
	car := mod + 0x03
	ch := o.chip

	// determine register value modifications
	modVol := volume.Volume(1)
	if (o.reg.RegC0 & 1) != 0 {
		// not additive
		modVol = carVol
	}

	var regA0, regB0 uint8
	if o.keyOn {
		freq, block := o.periodToFreqBlock(period, o.c2spd)
		regA0, regB0 = o.freqBlockToRegA0B0(freq, block)
		regB0 |= 0x20 // key on bit
	}

	// send the voice details out to the chip
	ch.WriteReg(0x20|mod, o.reg.Mod.Reg20)
	ch.WriteReg(0x40|mod, o.calc40(o.reg.Mod.Reg40, modVol))
	ch.WriteReg(0x60|mod, o.reg.Mod.Reg60)
	ch.WriteReg(0x80|mod, o.reg.Mod.Reg80)
	ch.WriteReg(0xE0|mod, o.reg.Mod.RegE0)

	ch.WriteReg(0xA0|index, regA0)

	ch.WriteReg(0x20|car, o.reg.Car.Reg20)
	ch.WriteReg(0x40|car, o.calc40(o.reg.Car.Reg40, carVol))
	ch.WriteReg(0x60|car, o.reg.Car.Reg60)
	ch.WriteReg(0x80|car, o.reg.Car.Reg80)
	ch.WriteReg(0xE0|car, o.reg.Car.RegE0)

	ch.WriteReg(0xC0|index, o.reg.RegC0)

	ch.WriteReg(0xB0|index, regB0)
}

// twoOperatorMelodic
var twoOperatorMelodic = [...]uint32{
	0x00, 0x01, 0x02, 0x08, 0x09, 0x0A, 0x10, 0x11, 0x12,
	0x100, 0x101, 0x102, 0x108, 0x109, 0x10A, 0x110, 0x111, 0x112,
}

func (o *OPL2) getChannelIndex(channelIdx int) uint32 {
	return twoOperatorMelodic[channelIdx%18]
}

func (o *OPL2) calc40(reg40 uint8, vol volume.Volume) uint8 {
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

func (o *OPL2) periodToFreqBlock(period note.Period, c2spd note.C2SPD) (uint16, uint8) {
	modFreq := period.GetFrequency()
	freq := float64(c2spd) * float64(modFreq) / 261625

	return o.freqToFnumBlock(freq)
}

func (o *OPL2) freqBlockToRegA0B0(freq uint16, block uint8) (uint8, uint8) {
	regA0 := uint8(freq)
	regB0 := uint8(uint16(freq)>>8) & 0x03
	regB0 |= (block & 0x07) << 3
	return regA0, regB0
}

func (o *OPL2) freqToFnumBlock(freq float64) (uint16, uint8) {
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
