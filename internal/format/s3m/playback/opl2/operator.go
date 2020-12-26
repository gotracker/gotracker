package opl2

// This file is a Pure Go conversion of dbopl.h/.cpp

/*
 *  Copyright (C) 2002-2013  The DOSBox Team
 *
 *  This program is free software; you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation; either version 2 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program; if not, write to the Free Software
 *  Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
 */

/*
	DOSBox implementation of a combined Yamaha YMF262 and Yamaha YM3812 emulator.
	Enabling the opl3 bit will switch the emulator to stereo opl3 output instead of regular mono opl2
	Except for the table generation it's all integer math
	Can choose different types of generators, using muls and bigger tables, try different ones for slower platforms
	The generation was based on the MAME implementation but tried to have it use less memory and be faster in general
	MAME uses much bigger envelope tables and this will be the biggest cause of it sounding different at times

	//TODO Don't delay first operator 1 sample in opl3 mode
	//TODO Maybe not use class method pointers but a regular function pointers with operator as first parameter
	//TODO Fix panning for the Percussion channels, would any opl3 player use it and actually really change it though?
	//TODO Check if having the same accuracy in all frequency multipliers sounds better or not

	//DUNNO Keyon in 4op, switch to 2op without keyoff.
*/

//Masks for operator 20 values
const (
	cMaskKSR     = 0x10
	cMaskSustain = 0x20
	cMaskVibrato = 0x40
	cMaskTremolo = 0x80
)

// OperatorState is a state of the operator's envelope
type OperatorState uint8

const (
	// OperatorStateOff is the OFF state of the operator envelope
	OperatorStateOff = OperatorState(iota)
	// OperatorStateRelease is the RELEASE state of the operator envelope
	OperatorStateRelease
	// OperatorStateSustain is the SUSTAIN state of the operator envelope
	OperatorStateSustain
	// OperatorStateDecay is the DECAY state of the operator envelope
	OperatorStateDecay
	// OperatorStateAttack is the ATTACK state of the operator envelope
	OperatorStateAttack
)

type volumeHandler func() int

// Operator is an OPL2/3 channel operator
type Operator struct {
	volHandler volumeHandler

	waveHandler waveHandler //Routine that generate a wave

	waveBase  []int16
	waveMask  int
	waveStart int

	waveIndex   int //WAVE_BITS shifted counter of the frequency index
	waveAdd     int //The base frequency without vibrato
	waveCurrent int //waveAdd + vibratao

	chanData     uint32 //Frequency/octave and derived data coming from whatever channel controls this
	freqMul      uint32 //Scale channel frequency with this, TODO maybe remove?
	vibrato      uint32 //Scaled up vibrato strength
	sustainLevel int32  //When stopping at sustain level stop here
	totalLevel   int32  //totalLevel is added to every generated volume
	currentLevel int    //totalLevel + tremolo
	volume       int32  //The currently active volume

	attackAdd  uint32 //Timers for the different states of the envelope
	decayAdd   uint32
	releaseAdd uint32
	rateIndex  uint32 //Current position of the evenlope

	rateZero uint8 //int for the different states of the envelope having no changes
	keyOn    uint8 //Bitmask of different values that can generate keyon
	//Registers, also used to check for changes
	reg20, reg40, reg60, reg80, regE0 uint8
	//Active part of the envelope we're in
	state OperatorState
	//0xff when tremolo is enabled
	tremoloMask uint8
	//Strength of the vibrato
	vibStrength uint8
	//Keep track of the calculated KSR so we can check for changes
	ksr uint8
}

// NewOperator creates a new OPL2/3 channel operator
func NewOperator() *Operator {
	o := Operator{}
	o.SetupOperator()

	return &o
}

// SetupOperator sets up a channel operator to defaults
func (o *Operator) SetupOperator() {
	o.SetState(OperatorStateOff)
	o.rateZero = (1 << OperatorStateOff)
	o.sustainLevel = cEnvMax
	o.currentLevel = cEnvMax
	o.totalLevel = cEnvMax
	o.volume = cEnvMax
}

// UpdateAttack updates the attack rate on the envelope
//We zero out when rate == 0
func (o *Operator) UpdateAttack(chip *Chip) {
	rate := uint8(o.reg60 >> 4)
	if rate != 0 {
		val := uint8((rate << 2) + o.ksr)
		o.attackAdd = chip.attackRates[val]
		o.rateZero &^= uint8(1 << OperatorStateAttack)
	} else {
		o.attackAdd = 0
		o.rateZero |= (1 << OperatorStateAttack)
	}
}

// UpdateDecay updates the decay rate on the envelope
func (o *Operator) UpdateDecay(chip *Chip) {
	rate := uint8(o.reg60 & 0xf)
	if rate != 0 {
		val := uint8((rate << 2) + o.ksr)
		o.decayAdd = chip.linearRates[val]
		o.rateZero &^= uint8(1 << OperatorStateDecay)
	} else {
		o.decayAdd = 0
		o.rateZero |= (1 << OperatorStateDecay)
	}
}

// UpdateRelease updates the release rate on the envelope
func (o *Operator) UpdateRelease(chip *Chip) {
	rate := uint8(o.reg80 & 0xf)
	if rate != 0 {
		val := uint8((rate << 2) + o.ksr)
		o.releaseAdd = chip.linearRates[val]
		o.rateZero &^= uint8(1 << OperatorStateRelease)
		if (o.reg20 & cMaskSustain) == 0 {
			o.rateZero &^= uint8(1 << OperatorStateSustain)
		}
	} else {
		o.rateZero |= (1 << OperatorStateRelease)
		o.releaseAdd = 0
		if (o.reg20 & cMaskSustain) == 0 {
			o.rateZero |= (1 << OperatorStateSustain)
		}
	}
}

// UpdateAttenuation updates the attenuation on the operator
func (o *Operator) UpdateAttenuation() {
	base := o.chanData >> cShiftKSLBase
	kslBase := uint8(base)
	tl := int32(o.reg40) & 0x3f
	kslShift := cKslShiftTable[o.reg40>>6]
	//Make sure the attenuation goes to the right bits
	o.totalLevel = tl << (cEnvBits - 7) //Total level goes 2 bits below max
	baseShift := int32(kslBase) << cEnvExtra
	o.totalLevel += baseShift >> kslShift
}

// UpdateFrequency updates the frequency on the operator
func (o *Operator) UpdateFrequency() {
	freq := uint32(o.chanData & ((1 << 10) - 1))
	block := uint32((o.chanData >> 10) & 0xff)
	if cWavePrecision != 0 {
		block = 7 - block
		o.waveAdd = int(freq*o.freqMul) >> block
	} else {
		o.waveAdd = int(freq*o.freqMul) << block
	}
	if (o.reg20 & cMaskVibrato) != 0 {
		o.vibStrength = (uint8)(freq >> 7)

		if cWavePrecision != 0 {
			o.vibrato = (uint32(o.vibStrength) * o.freqMul) >> block
		} else {
			o.vibrato = (uint32(o.vibStrength) << block) * o.freqMul
		}
	} else {
		o.vibStrength = 0
		o.vibrato = 0
	}
}

// UpdateRates updates the envelope rates on the operator
func (o *Operator) UpdateRates(chip *Chip) {
	//Mame seems to reverse this where enabling ksr actually lowers
	//the rate, but pdf manuals says otherwise?
	newKsr := uint8((uint8)((o.chanData >> cShiftKeyCode) & 0xff))
	if (o.reg20 & cMaskKSR) == 0 {
		newKsr >>= 2
	}
	if o.ksr == newKsr {
		return
	}
	o.ksr = newKsr
	o.UpdateAttack(chip)
	o.UpdateDecay(chip)
	o.UpdateRelease(chip)
}

// RateForward increments the operator's internal indexes
func (o *Operator) RateForward(add uint32) int32 {
	o.rateIndex += add
	ret := int32(o.rateIndex >> cRateSh)
	o.rateIndex = o.rateIndex & cRateMask
	return ret
}

// ForwardVolume updates the operator's current volume
func (o *Operator) ForwardVolume() int {
	return o.currentLevel + o.volHandler()
}

// ForwardWave updates the operator's current waveform
func (o *Operator) ForwardWave() uint {
	o.waveIndex += o.waveCurrent
	return uint(o.waveIndex) >> cWaveSh
}

// Write20 writes data to register 0x20 on the operator
func (o *Operator) Write20(chip *Chip, val uint8) {
	change := uint8((o.reg20 ^ val))
	if change == 0 {
		return
	}
	o.reg20 = val
	//Shift the tremolo bit over the entire register, saved a branch, YES!
	o.tremoloMask = val >> 7
	o.tremoloMask &^= uint8((1 << cEnvExtra) - 1)
	//Update specific features based on changes
	if (change & cMaskKSR) != 0 {
		o.UpdateRates(chip)
	}
	//With sustain enable the volume doesn't change
	if (o.reg20&cMaskSustain) != 0 || o.releaseAdd == 0 {
		o.rateZero |= (1 << OperatorStateSustain)
	} else {
		o.rateZero &^= uint8(1 << OperatorStateSustain)
	}
	//Frequency multiplier or vibrato changed
	if (change & (0xf | cMaskVibrato)) != 0 {
		o.freqMul = chip.freqMul[val&0xf]
		o.UpdateFrequency()
	}
}

// Write40 writes data to register 0x40 on the operator
func (o *Operator) Write40(chip *Chip, val uint8) {
	if (o.reg40 ^ val) == 0 {
		return
	}
	o.reg40 = val
	o.UpdateAttenuation()
}

// Write60 writes data to register 0x60 on the operator
func (o *Operator) Write60(chip *Chip, val uint8) {
	change := uint8(o.reg60 ^ val)
	o.reg60 = val
	if (change & 0x0f) != 0 {
		o.UpdateDecay(chip)
	}
	if (change & 0xf0) != 0 {
		o.UpdateAttack(chip)
	}
}

// Write80 writes data to register 0x80 on the operator
func (o *Operator) Write80(chip *Chip, val uint8) {
	change := uint8((o.reg80 ^ val))
	if change == 0 {
		return
	}
	o.reg80 = val
	sustain := uint8(val >> 4)
	//Turn 0xf into 0x1f
	sustain |= (sustain + 1) & 0x10
	o.sustainLevel = int32(sustain) << (cEnvBits - 5)
	if (change & 0x0f) != 0 {
		o.UpdateRelease(chip)
	}
}

// WriteE0 writes data to register 0xE0 on the operator
func (o *Operator) WriteE0(chip *Chip, val uint8) {
	if (o.regE0 ^ val) != 0 {
		return
	}
	//in opl3 mode you can always selet 7 waveforms regardless of waveformselect
	waveForm := uint8(val & (uint8(0x3&chip.waveFormMask) | (0x7 & uint8(chip.opl3Active))))
	o.regE0 = val
	if cDBOPLWave == cWaveHandler {
		o.waveHandler = waveHandlerTable[waveForm]
	} else {
		o.waveBase = cWaveTable[cWaveBaseTable[waveForm]:]
		o.waveStart = int(cWaveStartTable[waveForm]) << cWaveSh
		o.waveMask = int(cWaveMaskTable[waveForm])
	}
}

// SetState sets the current operator envelope state
func (o *Operator) SetState(s OperatorState) {
	o.state = s
	switch s {
	default:
		o.volHandler = o.volHandlerOFF
	case OperatorStateRelease:
		o.volHandler = o.volHandlerRELEASE
	case OperatorStateSustain:
		o.volHandler = o.volHandlerSUSTAIN
	case OperatorStateDecay:
		o.volHandler = o.volHandlerDECAY
	case OperatorStateAttack:
		o.volHandler = o.volHandlerATTACK
	}
}

func (o *Operator) volHandlerOFF() int {
	return cEnvMax
}

func (o *Operator) volHandlerRELEASE() int {
	vol := o.volume
	vol += o.RateForward(o.releaseAdd)
	if vol >= cEnvMax {
		o.volume = cEnvMax
		o.SetState(OperatorStateOff)
		return cEnvMax
	}
	o.volume = vol
	return int(vol)
}

func (o *Operator) volHandlerSUSTAIN() int {
	vol := o.volume
	if (o.reg20 & cMaskSustain) != 0 {
		return int(vol)
	}
	//In sustain phase, but not sustaining, do regular release
	o.SetState(OperatorStateRelease)
	return o.volHandlerRELEASE()
}

func (o *Operator) volHandlerDECAY() int {
	vol := o.volume
	vol += o.RateForward(o.decayAdd)
	if vol >= o.sustainLevel {
		//Check if we didn't overshoot max attenuation, then just go off
		if vol >= cEnvMax {
			o.volume = cEnvMax
			o.SetState(OperatorStateOff)
			return cEnvMax
		}
		//Continue as sustain
		o.rateIndex = 0
		o.SetState(OperatorStateSustain)
	}
	o.volume = vol
	return int(vol)
}

func (o *Operator) volHandlerATTACK() int {
	vol := o.volume
	change := o.RateForward(o.attackAdd)
	if change == 0 {
		return int(vol)
	}
	evol := ^vol
	evol *= change
	evol >>= 3
	vol += evol
	if vol < cEnvMin {
		o.volume = cEnvMin
		o.rateIndex = 0
		o.SetState(OperatorStateDecay)
		return cEnvMin
	}
	o.volume = vol
	return int(vol)
}

// Silent returns true if the operator is currently silent
func (o *Operator) Silent() bool {
	if !envSilent(int(o.totalLevel + o.volume)) {
		return false
	}
	if (o.rateZero & (1 << o.state)) == 0 {
		return false
	}
	return true
}

// Prepare prepares the operator's data
func (o *Operator) Prepare(chip *Chip) {
	o.currentLevel = int(o.totalLevel) + int(chip.tremoloValue&o.tremoloMask)
	o.waveCurrent = o.waveAdd
	if (o.vibStrength >> chip.vibratoShift) != 0 {
		add := int(o.vibrato) >> chip.vibratoShift
		//Sign extend over the shift value
		neg := int(chip.vibratoSign)
		//Negate the add with -1 or 0
		add ^= neg
		add -= neg
		o.waveCurrent += add
	}
}

// KeyOn updates the key-on state of the operator to true
func (o *Operator) KeyOn(mask uint8) {
	if o.keyOn == 0 {
		//Restart the frequency generator
		if cDBOPLWave > cWaveHandler {
			o.waveIndex = o.waveStart
		} else {
			o.waveIndex = 0
		}
		o.rateIndex = 0
		o.SetState(OperatorStateAttack)
	}
	o.keyOn |= mask
}

// KeyOff updates the key-on state of the operator to false
func (o *Operator) KeyOff(mask uint8) {
	o.keyOn &^= mask
	if o.keyOn == 0 {
		if o.state != OperatorStateOff {
			o.SetState(OperatorStateRelease)
		}
	}
}

// GetWave gets the current waveform of the operator
func (o *Operator) GetWave(index uint, vol int) int {
	if cDBOPLWave == cWaveHandler {
		return o.waveHandler(index, vol<<(3-cEnvExtra))
	} else if cDBOPLWave == cWaveTableMul {
		wb := o.waveBase[index&uint(o.waveMask)]
		base := int(wb)
		mul := int(cMulTable[vol>>cEnvExtra])
		val := (base * mul) >> cMulSh
		return val
	} else if cDBOPLWave == cWaveTableLog {
		wave := int32(o.waveBase[index&uint(o.waveMask)])
		total := uint32(int(wave&0x7fff) + vol<<(3-cEnvExtra))
		sig := int32(cExpTable[total&0xff])
		exp := total >> 8
		neg := int32(0)
		if wave < 0 {
			neg = -1
		}
		return int((sig^neg)-neg) >> exp
	} else {
		panic("No valid wave routine")
	}
}

// GetSample gets the current waveform of the operator, as a sample
func (o *Operator) GetSample(modulation int) int {
	vol := o.ForwardVolume()
	if envSilent(int(vol)) {
		//Simply forward the wave
		o.waveIndex += o.waveCurrent
		return 0
	}
	index := o.ForwardWave()
	index = uint(int(index) + modulation)
	return o.GetWave(index, vol)
}
