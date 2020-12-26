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
	MASK_KSR     = 0x10
	MASK_SUSTAIN = 0x20
	MASK_VIBRATO = 0x40
	MASK_TREMOLO = 0x80
)

type State uint8

const (
	OFF = State(iota)
	RELEASE
	SUSTAIN
	DECAY
	ATTACK
)

type VolumeHandler func() Bits

type Operator struct {
	volHandler VolumeHandler

	waveHandler WaveHandler //Routine that generate a wave

	waveBase  []int16
	waveMask  uint32
	waveStart uint32

	waveIndex   uint32 //WAVE_BITS shifted counter of the frequency index
	waveAdd     uint32 //The base frequency without vibrato
	waveCurrent uint32 //waveAdd + vibratao

	chanData     uint32 //Frequency/octave and derived data coming from whatever channel controls this
	freqMul      uint32 //Scale channel frequency with this, TODO maybe remove?
	vibrato      uint32 //Scaled up vibrato strength
	sustainLevel int32  //When stopping at sustain level stop here
	totalLevel   int32  //totalLevel is added to every generated volume
	currentLevel uint32 //totalLevel + tremolo
	volume       int32  //The currently active volume

	attackAdd  uint32 //Timers for the different states of the envelope
	decayAdd   uint32
	releaseAdd uint32
	rateIndex  uint32 //Current position of the evenlope

	rateZero uint8 //Bits for the different states of the envelope having no changes
	keyOn    uint8 //Bitmask of different values that can generate keyon
	//Registers, also used to check for changes
	reg20, reg40, reg60, reg80, regE0 uint8
	//Active part of the envelope we're in
	state State
	//0xff when tremolo is enabled
	tremoloMask uint8
	//Strength of the vibrato
	vibStrength uint8
	//Keep track of the calculated KSR so we can check for changes
	ksr uint8
}

func NewOperator() *Operator {
	o := Operator{}
	o.SetupOperator()

	return &o
}

func (o *Operator) SetupOperator() {
	o.SetState(OFF)
	o.rateZero = (1 << OFF)
	o.sustainLevel = ENV_MAX
	o.currentLevel = ENV_MAX
	o.totalLevel = ENV_MAX
	o.volume = ENV_MAX
}

//We zero out when rate == 0
func (o *Operator) UpdateAttack(chip *Chip) {
	rate := uint8(o.reg60 >> 4)
	if rate != 0 {
		val := uint8((rate << 2) + o.ksr)
		o.attackAdd = chip.attackRates[val]
		o.rateZero &= ^uint8(1 << ATTACK)
	} else {
		o.attackAdd = 0
		o.rateZero |= (1 << ATTACK)
	}
}
func (o *Operator) UpdateDecay(chip *Chip) {
	rate := uint8(o.reg60 & 0xf)
	if rate != 0 {
		val := uint8((rate << 2) + o.ksr)
		o.decayAdd = chip.linearRates[val]
		o.rateZero &= ^uint8(1 << DECAY)
	} else {
		o.decayAdd = 0
		o.rateZero |= (1 << DECAY)
	}
}
func (o *Operator) UpdateRelease(chip *Chip) {
	rate := uint8(o.reg80 & 0xf)
	if rate != 0 {
		val := uint8((rate << 2) + o.ksr)
		o.releaseAdd = chip.linearRates[val]
		o.rateZero &= ^uint8(1 << RELEASE)
		if (o.reg20 & MASK_SUSTAIN) == 0 {
			o.rateZero &= ^uint8(1 << SUSTAIN)
		}
	} else {
		o.rateZero |= (1 << RELEASE)
		o.releaseAdd = 0
		if (o.reg20 & MASK_SUSTAIN) == 0 {
			o.rateZero |= (1 << SUSTAIN)
		}
	}
}

func (o *Operator) UpdateAttenuation() {
	kslBase := uint8((uint8)((o.chanData >> SHIFT_KSLBASE) & 0xff))
	tl := uint32(o.reg40 & 0x3f)
	kslShift := uint8(KslShiftTable[o.reg40>>6])
	//Make sure the attenuation goes to the right bits
	o.totalLevel = int32(tl << (ENV_BITS - 7)) //Total level goes 2 bits below max
	o.totalLevel += int32((kslBase << ENV_EXTRA) >> kslShift)
}

func (o *Operator) UpdateFrequency() {
	freq := uint32(o.chanData & ((1 << 10) - 1))
	block := uint32((o.chanData >> 10) & 0xff)
	if WAVE_PRECISION != 0 {
		block = 7 - block
		o.waveAdd = (freq * o.freqMul) >> block
	} else {
		o.waveAdd = (freq << block) * o.freqMul
	}
	if (o.reg20 & MASK_VIBRATO) != 0 {
		o.vibStrength = (uint8)(freq >> 7)

		if WAVE_PRECISION != 0 {
			o.vibrato = (uint32(o.vibStrength) * o.freqMul) >> block
		} else {
			o.vibrato = (uint32(o.vibStrength) << block) * o.freqMul
		}
	} else {
		o.vibStrength = 0
		o.vibrato = 0
	}
}

func (o *Operator) UpdateRates(chip *Chip) {
	//Mame seems to reverse this where enabling ksr actually lowers
	//the rate, but pdf manuals says otherwise?
	newKsr := uint8((uint8)((o.chanData >> SHIFT_KEYCODE) & 0xff))
	if (o.reg20 & MASK_KSR) == 0 {
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

func (o *Operator) RateForward(add uint32) int32 {
	o.rateIndex += add
	ret := int32(o.rateIndex >> RATE_SH)
	o.rateIndex = o.rateIndex & RATE_MASK
	return ret
}

func (o *Operator) TemplateVolume(yes State) Bits {
	vol := int32(o.volume)
	var change int32
	switch yes {
	case OFF:
		return ENV_MAX
	case ATTACK:
		change = o.RateForward(o.attackAdd)
		if change == 0 {
			return Bits(vol)
		}
		vol += ((^vol) * change) >> 3
		if vol < ENV_MIN {
			o.volume = ENV_MIN
			o.rateIndex = 0
			o.SetState(DECAY)
			return ENV_MIN
		}
	case DECAY:
		vol += o.RateForward(o.decayAdd)
		if vol >= o.sustainLevel {
			//Check if we didn't overshoot max attenuation, then just go off
			if vol >= ENV_MAX {
				o.volume = ENV_MAX
				o.SetState(OFF)
				return ENV_MAX
			}
			//Continue as sustain
			o.rateIndex = 0
			o.SetState(SUSTAIN)
		}
	case SUSTAIN:
		if (o.reg20 & MASK_SUSTAIN) != 0 {
			return Bits(vol)
		}
		//In sustain phase, but not sustaining, do regular release
		fallthrough
	case RELEASE:
		vol += o.RateForward(o.releaseAdd)
		if vol >= ENV_MAX {
			o.volume = ENV_MAX
			o.SetState(OFF)
			return ENV_MAX
		}
	}
	o.volume = vol
	return Bits(vol)
}

func (o *Operator) ForwardVolume() Bitu {
	return Bitu(Bits(o.currentLevel) + o.volHandler())
}

func (o *Operator) ForwardWave() Bitu {
	o.waveIndex += o.waveCurrent
	return Bitu(o.waveIndex) >> WAVE_SH
}

func (o *Operator) Write20(chip *Chip, val uint8) {
	change := uint8((o.reg20 ^ val))
	if change == 0 {
		return
	}
	o.reg20 = val
	//Shift the tremolo bit over the entire register, saved a branch, YES!
	o.tremoloMask = val >> 7
	o.tremoloMask &= ^uint8((1 << ENV_EXTRA) - 1)
	//Update specific features based on changes
	if (change & MASK_KSR) != 0 {
		o.UpdateRates(chip)
	}
	//With sustain enable the volume doesn't change
	if (o.reg20&MASK_SUSTAIN) != 0 || o.releaseAdd == 0 {
		o.rateZero |= (1 << SUSTAIN)
	} else {
		o.rateZero &= ^uint8(1 << SUSTAIN)
	}
	//Frequency multiplier or vibrato changed
	if (change & (0xf | MASK_VIBRATO)) != 0 {
		o.freqMul = chip.freqMul[val&0xf]
		o.UpdateFrequency()
	}
}

func (o *Operator) Write40(chip *Chip, val uint8) {
	if (o.reg40 ^ val) == 0 {
		return
	}
	o.reg40 = val
	o.UpdateAttenuation()
}

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

func (o *Operator) Write80(chip *Chip, val uint8) {
	change := uint8((o.reg80 ^ val))
	if change == 0 {
		return
	}
	o.reg80 = val
	sustain := uint8(val >> 4)
	//Turn 0xf into 0x1f
	sustain |= (sustain + 1) & 0x10
	o.sustainLevel = int32(sustain) << (ENV_BITS - 5)
	if (change & 0x0f) != 0 {
		o.UpdateRelease(chip)
	}
}

func (o *Operator) WriteE0(chip *Chip, val uint8) {
	if (o.regE0 ^ val) != 0 {
		return
	}
	//in opl3 mode you can always selet 7 waveforms regardless of waveformselect
	waveForm := uint8(val & (uint8(0x3&chip.waveFormMask) | (0x7 & uint8(chip.opl3Active))))
	o.regE0 = val
	if DBOPL_WAVE == WAVE_HANDLER {
		o.waveHandler = WaveHandlerTable[waveForm]
	} else {
		o.waveBase = WaveTable[WaveBaseTable[waveForm]:]
		o.waveStart = uint32(WaveStartTable[waveForm]) << WAVE_SH
		o.waveMask = uint32(WaveMaskTable[waveForm])
	}
}

func (o *Operator) SetState(s State) {
	o.state = s
	o.volHandler = func() Bits {
		return o.TemplateVolume(s)
	}
}

func (o *Operator) Silent() bool {
	if !ENV_SILENT(int(o.totalLevel + o.volume)) {
		return false
	}
	if (o.rateZero & (1 << o.state)) == 0 {
		return false
	}
	return true
}

func (o *Operator) Prepare(chip *Chip) {
	o.currentLevel = uint32(o.totalLevel) + uint32(chip.tremoloValue&o.tremoloMask)
	o.waveCurrent = o.waveAdd
	if (o.vibStrength >> chip.vibratoShift) != 0 {
		add := int32(o.vibrato >> chip.vibratoShift)
		//Sign extend over the shift value
		neg := int32(chip.vibratoSign)
		//Negate the add with -1 or 0
		add = (add ^ neg) - neg
		o.waveCurrent = uint32(int32(o.waveCurrent) + add)
	}
}

func (o *Operator) KeyOn(mask uint8) {
	if o.keyOn == 0 {
		//Restart the frequency generator
		if DBOPL_WAVE > WAVE_HANDLER {
			o.waveIndex = o.waveStart
		} else {
			o.waveIndex = 0
		}
		o.rateIndex = 0
		o.SetState(ATTACK)
	}
	o.keyOn |= mask
}

func (o *Operator) KeyOff(mask uint8) {
	o.keyOn &= ^mask
	if o.keyOn == 0 {
		if o.state != OFF {
			o.SetState(RELEASE)
		}
	}
}

func (o *Operator) GetWave(index Bitu, vol Bitu) Bits {
	if DBOPL_WAVE == WAVE_HANDLER {
		return o.waveHandler(index, vol<<(3-ENV_EXTRA))
	} else if DBOPL_WAVE == WAVE_TABLEMUL {
		return Bits((uint32(o.waveBase[index&Bitu(o.waveMask)]) * uint32(MulTable[vol>>ENV_EXTRA])) >> MUL_SH)
	} else if DBOPL_WAVE == WAVE_TABLELOG {
		wave := int32(o.waveBase[index&Bitu(o.waveMask)])
		total := uint32(Bitu(wave&0x7fff) + vol<<(3-ENV_EXTRA))
		sig := int32(ExpTable[total&0xff])
		exp := uint32(total >> 8)
		neg := int32(wave >> 16)
		return Bits((sig^neg)-neg) >> exp
	} else {
		panic("No valid wave routine")
	}
}

func (o *Operator) GetSample(modulation Bits) Bits {
	vol := o.ForwardVolume()
	if ENV_SILENT(int(vol)) {
		//Simply forward the wave
		o.waveIndex += o.waveCurrent
		return 0
	} else {
		index := Bitu(o.ForwardWave())
		index = Bitu(Bits(index) + modulation)
		return o.GetWave(index, vol)
	}
}
