package opl2

// This file is a Pure Go conversion of dbopl.h/.cpp, slimmed down to just a single channel's
// operator. It technically has diverged significantly enough from the original programming
// that it probably doesn't need the below license notice, but to be a good citizen, I've kept
// it here.

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
	maskKSR     = 0x10
	maskSustain = 0x20
	maskVibrato = 0x40
	maskTremolo = 0x80
)

type envelopeState uint8

const (
	envStateOff = envelopeState(iota)
	envStateRelease
	envStateSustain
	envStateDecay
	envStateAttack
)

type volumeHandler func() int
type waveHandler func(uint, uint) int

type SingleChannelOp struct {
	volHandler volumeHandler

	waveHandler waveHandler //Routine that generate a wave

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

	rateZero uint8 //int for the different states of the envelope having no changes
	keyOn    uint8 //Bitmask of different values that can generate keyon
	//Registers, also used to check for changes
	reg20, reg40, reg60, reg80 uint8
	waveForm                   uint8
	//Active part of the envelope we're in
	state envelopeState
	//0xff when tremolo is enabled
	tremoloMask uint8
	//Strength of the vibrato
	vibStrength uint8
	//Keep track of the calculated KSR so we can check for changes
	ksr uint8
}

func NewSingleOperator() *SingleChannelOp {
	o := SingleChannelOp{}
	o.SetupOperator()

	return &o
}

func (o *SingleChannelOp) SetupOperator() {
	o.SetState(envStateOff)
	o.rateZero = (1 << envStateOff)
	o.sustainLevel = ENV_MAX
	o.currentLevel = ENV_MAX
	o.totalLevel = ENV_MAX
	o.volume = ENV_MAX
}

//We zero out when rate == 0
func (o *SingleChannelOp) UpdateAttack(ch *SingleChannel) {
	rate := uint8(o.reg60 >> 4)
	if rate != 0 {
		val := uint8((rate << 2) + o.ksr)
		o.attackAdd = ch.attackRates[val]
		o.rateZero &= ^uint8(1 << envStateAttack)
	} else {
		o.attackAdd = 0
		o.rateZero |= (1 << envStateAttack)
	}
}
func (o *SingleChannelOp) UpdateDecay(ch *SingleChannel) {
	rate := uint8(o.reg60 & 0xf)
	if rate != 0 {
		val := uint8((rate << 2) + o.ksr)
		o.decayAdd = ch.linearRates[val]
		o.rateZero &= ^uint8(1 << envStateDecay)
	} else {
		o.decayAdd = 0
		o.rateZero |= (1 << envStateDecay)
	}
}
func (o *SingleChannelOp) UpdateRelease(ch *SingleChannel) {
	rate := uint8(o.reg80 & 0xf)
	if rate != 0 {
		val := uint8((rate << 2) + o.ksr)
		o.releaseAdd = ch.linearRates[val]
		o.rateZero &= ^uint8(1 << envStateRelease)
		if (o.reg20 & maskSustain) == 0 {
			o.rateZero &= ^uint8(1 << envStateSustain)
		}
	} else {
		o.rateZero |= (1 << envStateRelease)
		o.releaseAdd = 0
		if (o.reg20 & maskSustain) == 0 {
			o.rateZero |= (1 << envStateSustain)
		}
	}
}

//Shift strength for the ksl value determined by ksl strength
var kslShiftTable = [4]uint8{
	31, 1, 2, 0,
}

func (o *SingleChannelOp) UpdateAttenuation() {
	kslBase := uint8((uint8)((o.chanData >> shiftKSLBase) & 0xff))
	tl := uint32(o.reg40 & 0x3f)
	kslShift := uint8(kslShiftTable[o.reg40>>6])
	//Make sure the attenuation goes to the right bits
	o.totalLevel = int32(tl << (ENV_BITS - 7)) //Total level goes 2 bits below max
	o.totalLevel += int32((kslBase << ENV_EXTRA) >> kslShift)
}

func (o *SingleChannelOp) UpdateFrequency() {
	freq := uint32(o.chanData & ((1 << 10) - 1))
	block := uint32((o.chanData >> 10) & 0xff)
	if WAVE_PRECISION != 0 {
		block = 7 - block
		o.waveAdd = (freq * o.freqMul) >> block
	} else {
		o.waveAdd = (freq << block) * o.freqMul
	}
	if (o.reg20 & maskVibrato) != 0 {
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

func (o *SingleChannelOp) UpdateRates(ch *SingleChannel) {
	//Mame seems to reverse this where enabling ksr actually lowers
	//the rate, but pdf manuals says otherwise?
	newKsr := uint8((uint8)((o.chanData >> shiftKSLCode) & 0xff))
	if (o.reg20 & maskKSR) == 0 {
		newKsr >>= 2
	}
	if o.ksr == newKsr {
		return
	}
	o.ksr = newKsr
	o.UpdateAttack(ch)
	o.UpdateDecay(ch)
	o.UpdateRelease(ch)
}

func (o *SingleChannelOp) RateForward(add uint32) int32 {
	o.rateIndex += add
	ret := int32(o.rateIndex >> RATE_SH)
	o.rateIndex = o.rateIndex & RATE_MASK
	return ret
}

func (o *SingleChannelOp) TemplateVolume(yes envelopeState) int {
	vol := int32(o.volume)
	var change int32
	switch yes {
	case envStateOff:
		return ENV_MAX
	case envStateAttack:
		change = o.RateForward(o.attackAdd)
		if change == 0 {
			return int(vol)
		}
		vol += ((^vol) * change) >> 3
		if vol < ENV_MIN {
			o.volume = ENV_MIN
			o.rateIndex = 0
			o.SetState(envStateDecay)
			return ENV_MIN
		}
	case envStateDecay:
		vol += o.RateForward(o.decayAdd)
		if vol >= o.sustainLevel {
			//Check if we didn't overshoot max attenuation, then just go off
			if vol >= ENV_MAX {
				o.volume = ENV_MAX
				o.SetState(envStateOff)
				return ENV_MAX
			}
			//Continue as sustain
			o.rateIndex = 0
			o.SetState(envStateSustain)
		}
	case envStateSustain:
		if (o.reg20 & maskSustain) != 0 {
			return int(vol)
		}
		//In sustain phase, but not sustaining, do regular release
		fallthrough
	case envStateRelease:
		vol += o.RateForward(o.releaseAdd)
		if vol >= ENV_MAX {
			o.volume = ENV_MAX
			o.SetState(envStateOff)
			return ENV_MAX
		}
	}
	o.volume = vol
	return int(vol)
}

func (o *SingleChannelOp) ForwardVolume() uint {
	return uint(int(o.currentLevel) + o.volHandler())
}

func (o *SingleChannelOp) ForwardWave() uint {
	o.waveIndex += o.waveCurrent
	return uint(o.waveIndex) >> WAVE_SH
}

func (o *SingleChannelOp) Write20(ch *SingleChannel, val uint8) {
	change := uint8((o.reg20 ^ val))
	if change == 0 {
		return
	}
	o.reg20 = val
	//Shift the tremolo bit over the entire register, saved a branch, YES!
	o.tremoloMask = val >> 7
	o.tremoloMask &= ^uint8((1 << ENV_EXTRA) - 1)
	//Update specific features based on changes
	if (change & maskKSR) != 0 {
		o.UpdateRates(ch)
	}
	//With sustain enable the volume doesn't change
	if (o.reg20&maskSustain) != 0 || o.releaseAdd == 0 {
		o.rateZero |= (1 << envStateSustain)
	} else {
		o.rateZero &= ^uint8(1 << envStateSustain)
	}
	//Frequency multiplier or vibrato changed
	if (change & (0xf | maskVibrato)) != 0 {
		o.freqMul = ch.freqMul[val&0xf]
		o.UpdateFrequency()
	}
}

func (o *SingleChannelOp) Write40(val uint8) {
	if (o.reg40 ^ val) == 0 {
		return
	}
	o.reg40 = val
	o.UpdateAttenuation()
}

func (o *SingleChannelOp) Write60(ch *SingleChannel, val uint8) {
	change := uint8(o.reg60 ^ val)
	o.reg60 = val
	if (change & 0x0f) != 0 {
		o.UpdateDecay(ch)
	}
	if (change & 0xf0) != 0 {
		o.UpdateAttack(ch)
	}
}

func (o *SingleChannelOp) Write80(ch *SingleChannel, val uint8) {
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
		o.UpdateRelease(ch)
	}
}

func (o *SingleChannelOp) WriteWaveForm(ch *SingleChannel, waveForm uint8) {
	waveForm &= ch.waveFormMask
	if (o.waveForm ^ waveForm) != 0 {
		return
	}
	o.waveForm = waveForm
	if DBOPL_WAVE == WAVE_HANDLER {
		o.waveHandler = waveHandlerTable[waveForm]
	} else {
		o.waveBase = waveTable[waveBaseTable[waveForm]:]
		o.waveStart = uint32(waveStartTable[waveForm]) << WAVE_SH
		o.waveMask = uint32(waveMaskTable[waveForm])
	}
}

func (o *SingleChannelOp) SetState(s envelopeState) {
	o.state = s
	o.volHandler = func() int {
		return o.TemplateVolume(s)
	}
}

func (o *SingleChannelOp) Silent() bool {
	if !ENV_SILENT(int(o.totalLevel + o.volume)) {
		return false
	}
	if (o.rateZero & (1 << o.state)) == 0 {
		return false
	}
	return true
}

func (o *SingleChannelOp) Prepare(ch *SingleChannel) {
	o.currentLevel = uint32(o.totalLevel) + uint32(ch.tremoloValue&o.tremoloMask)
	o.waveCurrent = o.waveAdd
	if (o.vibStrength >> ch.vibratoShift) != 0 {
		add := int32(o.vibrato >> ch.vibratoShift)
		//Sign extend over the shift value
		neg := int32(ch.vibratoSign)
		//Negate the add with -1 or 0
		add = (add ^ neg) - neg
		o.waveCurrent = uint32(int32(o.waveCurrent) + add)
	}
}

func (o *SingleChannelOp) KeyOn(mask uint8) {
	if o.keyOn == 0 {
		//Restart the frequency generator
		if DBOPL_WAVE > WAVE_HANDLER {
			o.waveIndex = o.waveStart
		} else {
			o.waveIndex = 0
		}
		o.rateIndex = 0
		o.SetState(envStateAttack)
	}
	o.keyOn |= mask
}

func (o *SingleChannelOp) KeyOff(mask uint8) {
	o.keyOn &= ^mask
	if o.keyOn == 0 {
		if o.state != envStateOff {
			o.SetState(envStateRelease)
		}
	}
}

func (o *SingleChannelOp) GetWave(index uint, vol uint) int {
	if DBOPL_WAVE == WAVE_HANDLER {
		return o.waveHandler(index, vol<<(3-ENV_EXTRA))
	} else if DBOPL_WAVE == WAVE_TABLEMUL {
		return int((uint32(o.waveBase[index&uint(o.waveMask)]) * uint32(mulTable[vol>>ENV_EXTRA])) >> MUL_SH)
	} else if DBOPL_WAVE == WAVE_TABLELOG {
		wave := int32(o.waveBase[index&uint(o.waveMask)])
		total := uint32(uint(wave&0x7fff) + vol<<(3-ENV_EXTRA))
		sig := int32(expTable[total&0xff])
		exp := uint32(total >> 8)
		neg := int32(wave >> 16)
		return int((sig^neg)-neg) >> exp
	} else {
		panic("No valid wave routine")
	}
}

func (o *SingleChannelOp) GetSample(modulation int) int {
	vol := o.ForwardVolume()
	if ENV_SILENT(int(vol)) {
		//Simply forward the wave
		o.waveIndex += o.waveCurrent
		return 0
	} else {
		index := uint(o.ForwardWave())
		index = uint(int(index) + modulation)
		return o.GetWave(index, vol)
	}
}
