package opl2

// This file is a Pure Go conversion of dbopl.h/.cpp, slimmed down to just a single channel
// It technically has diverged significantly enough from the original programming that it
// probably doesn't need the below license notice, but to be a good citizen, I've kept it
// here.

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
 *  along with c program; if not, write to the Free Software
 *  Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
 */

/*
	DOSBox implementation of a combined Yamaha YMF262 and Yamaha YM3812 emulator.
	Enabling the opl3 bit will switch the emulator to stereo opl3 output instead of regular mono opl2
	Except for the table generation it's all integer math
	Can choose different types of generators, using muls and bigger tables, try different ones for slower platforms
	The generation was based on the MAME implementation but tried to have it use less memory and be faster in general
	MAME uses much bigger envelope tables and c will be the biggest cause of it sounding different at times

	//TODO Don't delay first operator 1 sample in opl3 mode
	//TODO Maybe not use class method pointers but a regular function pointers with operator as first parameter
	//TODO Fix panning for the Percussion channels, would any opl3 player use it and actually really change it though?
	//TODO Check if having the same accuracy in all frequency multipliers sounds better or not

	//DUNNO Keyon in 4op, switch to 2op without keyoff.
*/

type SingleChannel struct {
	op [2]SingleChannelOp

	additiveSynthesis bool

	chanData uint32   //Frequency/octave and derived values
	old      [2]int32 //Old data for feedback

	feedback uint8 //Feedback shift
	keyOn    bool
	fnum     uint16
	block    uint8
	freqHi   uint8
	//This should correspond with reg104, bit 6 indicates a Percussion channel, bit 7 indicates a silent channel
	maskLeft  int8 //Sign extended values for both channel's panning
	maskRight int8

	// === from Chip ===
	reg08 uint8
	//This is used as the base counter for vibrato and tremolo
	lfoCounter uint32
	lfoAdd     uint32

	noiseCounter uint32
	noiseAdd     uint32
	noiseValue   uint32

	vibratoIndex    uint8
	tremoloIndex    uint8
	vibratoSign     int8
	vibratoShift    uint8
	tremoloValue    uint8
	vibratoStrength uint8
	tremoloStrength uint8
	//Mask for allowed wave forms
	waveFormMask uint8
	//Frequency scales for the different multiplications
	freqMul [16]uint32
	//Rates for decay and release for rate of this chip/op
	linearRates [76]uint32
	//Best match attack rates for the rate of this chip/op
	attackRates [76]uint32
}

func NewSingleChannel(rate uint32) *SingleChannel {
	c := SingleChannel{}
	c.SetupChannel(rate)
	return &c
}

const (
	shiftKSLBase = 16

	shiftKSLCode = 24
)

func (c *SingleChannel) SetChanData(data uint32) {
	change := c.chanData ^ data
	c.chanData = data
	c.op[0].chanData = data
	c.op[1].chanData = data
	//Since a frequency update triggered c, always update frequency
	c.op[0].UpdateFrequency()
	c.op[1].UpdateFrequency()
	if (change & (0xff << shiftKSLBase)) != 0 {
		c.op[0].UpdateAttenuation()
		c.op[1].UpdateAttenuation()
	}
	if (change & (0xff << shiftKSLCode)) != 0 {
		c.op[0].UpdateRates(c)
		c.op[1].UpdateRates(c)
	}
}

func (c *SingleChannel) UpdateFrequency() {
	//Extract the frequency int
	data := c.chanData & 0xffff
	kslBase := kslTable[data>>6]
	keyCode := (data & 0x1c00) >> 9
	if (c.reg08 & 0x40) != 0 {
		keyCode |= (data & 0x100) >> 8 // notesel == 1
	} else {
		keyCode |= (data & 0x200) >> 9 // notesel == 0
	}
	//Add the keycode and ksl into the highest int of chanData
	data |= (keyCode << shiftKSLCode) | (uint32(kslBase) << shiftKSLBase)
	c.SetChanData(data)
}

func (c *SingleChannel) SetFNum(fnum uint16, block uint8) {
	var regB0A0 uint16
	regB0A0 |= fnum & 0x3ff
	regB0A0 |= uint16(block&0x07) << 10
	// don't include keyOn bit
	change := uint(c.chanData ^ (uint32(regB0A0) & 0x7fff))
	if change != 0 {
		c.chanData ^= uint32(change)
		c.fnum = fnum
		c.block = block
		c.UpdateFrequency()
	}
}

func (c *SingleChannel) SetKeyOn(on bool) {
	if on == c.keyOn {
		return
	}
	c.keyOn = on
	if on {
		c.op[0].KeyOn(0x1)
		c.op[1].KeyOn(0x1)
	} else {
		c.op[0].KeyOff(0x1)
		c.op[1].KeyOff(0x1)
	}
}

func (c *SingleChannel) GetKeyOn() bool {
	return c.keyOn
}

func (c *SingleChannel) SetAdditiveSynthesis(on bool) {
	c.additiveSynthesis = on
}

func (c *SingleChannel) SetModulationFeedback(feedback uint8) {
	feedback &= 0x07
	if feedback != 0 {
		//We shift the input to the right 10 bit wave index value
		c.feedback = 9 - feedback
	} else {
		c.feedback = 31
	}
}

func (c *SingleChannel) ResetC0() {
	c.SetAdditiveSynthesis(false)
	c.SetModulationFeedback(0)
}

func (c *SingleChannel) BlockTemplate(samples uint32, output []int32) {
	if c.additiveSynthesis {
		if c.op[0].Silent() && c.op[1].Silent() {
			c.old[0] = 0
			c.old[1] = 0
			return
		}
	} else {
		if c.op[1].Silent() {
			c.old[0] = 0
			c.old[1] = 0
			return
		}
	}
	//Init the operators with the the current vibrato and tremolo values
	c.op[0].Prepare(c)
	c.op[1].Prepare(c)
	for i := uint(0); i < uint(samples); i++ {
		//Do unsigned shift so we can shift out all int but still stay in 10 bit range otherwise
		mod := int((c.old[0] + c.old[1]) >> c.feedback)
		c.old[0] = c.old[1]
		c.old[1] = int32(c.op[0].GetSample(mod))
		var sample int32
		out0 := int(c.old[0])
		if c.additiveSynthesis {
			sample = int32(out0 + c.op[1].GetSample(0))
		} else {
			sample = int32(c.op[1].GetSample(out0))
		}
		output[i] += sample
	}
}

func (c *SingleChannel) ForwardNoise() uint32 {
	c.noiseCounter += c.noiseAdd
	count := uint(c.noiseCounter >> bitsLFOShift)
	c.noiseCounter &= bitsWaveMask
	for ; count > 0; count-- {
		//Noise calculation from mame
		c.noiseValue ^= (0x800302) & (0 - (c.noiseValue & 1))
		c.noiseValue >>= 1
	}
	return c.noiseValue
}

func (c *SingleChannel) ForwardLFO(samples uint32) uint32 {
	//Current vibrato value, runs 4x slower than tremolo
	c.vibratoSign = (vibratoTable[c.vibratoIndex>>2]) >> 7
	c.vibratoShift = uint8(vibratoTable[c.vibratoIndex>>2]&7) + c.vibratoStrength
	c.tremoloValue = tremoloTable[c.tremoloIndex] >> c.tremoloStrength

	//Check hom many samples there can be done before the value changes
	todo := uint32(lfoMax) - c.lfoCounter
	count := uint32((todo + c.lfoAdd - 1) / c.lfoAdd)
	if count > samples {
		count = samples
		c.lfoCounter += count * c.lfoAdd
	} else {
		c.lfoCounter += count * c.lfoAdd
		c.lfoCounter &= uint32(lfoMax - 1)
		//Maximum of 7 vibrato value * 4
		c.vibratoIndex = (c.vibratoIndex + 1) & 31
		//Clip tremolo to the the table size
		if c.tremoloIndex+1 < TREMOLO_TABLE {
			c.tremoloIndex++
		} else {
			c.tremoloIndex = 0
		}
	}
	return count
}

func (c *SingleChannel) GenerateBlock2(total uint, output []int32) {
	outputIdx := uint(0)
	for total > 0 {
		samples := c.ForwardLFO(uint32(total))
		c.BlockTemplate(samples, output[outputIdx:])
		total -= uint(samples)
		outputIdx += uint(samples)
	}
}

func (c *SingleChannel) SetupChannel(rate uint32) {
	original := float64(OPLRATE)
	scale := original / float64(rate)

	c.feedback = 31
	c.maskLeft = -1
	c.maskRight = -1
	c.additiveSynthesis = false
	for i := range c.op {
		c.op[i].SetupOperator()
	}

	//Noise counter is run at the same precision as general waves
	c.noiseAdd = (uint32)(0.5 + scale*float64(uint32(1)<<bitsLFOShift))
	c.noiseCounter = 0
	c.noiseValue = 1 //Make sure it triggers the noise xor the first time
	//The low frequency oscillation counter
	//Every time his overflows vibrato and tremoloindex are increased
	c.lfoAdd = uint32(0.5 + scale*float64(uint32(1)<<bitsLFOShift))
	c.lfoCounter = 0
	c.vibratoIndex = 0
	c.tremoloIndex = 0

	//With higher octave this gets shifted up
	//-1 since the freqCreateTable = *2
	if WAVE_PRECISION != 0 {
		freqScale := float64(float64(1<<7) * scale * float64(uint(1)<<(bitsWaveShift-1-10)))
		for i := 0; i < 16; i++ {
			c.freqMul[i] = uint32(0.5 + freqScale*float64(freqCreateTable[i]))
		}
	} else {
		freqScale := uint32(0.5 + scale*float64(uint(1)<<(bitsWaveShift-1-10)))
		for i := 0; i < 16; i++ {
			c.freqMul[i] = freqScale * freqCreateTable[i]
		}
	}

	//-3 since the real envelope takes 8 steps to reach the single value we supply
	for i := uint8(0); i < 76; i++ {
		scaledIncrease := getScaledIncreaseEnvelope(i, scale)
		c.linearRates[i] = scaledIncrease
	}
	//Generate the best matching attack rate
	for i := uint8(0); i < 62; i++ {
		index, shift := envelopeSelect(i)
		//Original amount of samples the attack would take
		attackSamples := uint32(attackSamplesTable[index]) << shift
		originalSamples := uint32(float64(attackSamples) / scale)

		scaledIncrease := getScaledIncreaseEnvelope(i, scale)
		guessAdd := scaledIncrease
		bestAdd := guessAdd
		bestDiff := int32(1) << 30
	passesLoop:
		for passes := 0; passes < 16; passes++ {
			volume := int32(envelopeMax)
			samples := uint32(0)
			count := uint32(0)
			for volume > 0 && samples < originalSamples*2 {
				count += guessAdd
				change := count >> bitsRateShift
				count &= bitsRateMask
				if change != 0 { // less than 1 %
					volume += (^volume * int32(change)) >> 3
				}
				samples++
			}
			diff := int32(originalSamples) - int32(samples)
			lDiff := diff
			if lDiff < 0 {
				lDiff = -lDiff
			}
			//Init last on first pass
			if lDiff < bestDiff {
				bestDiff := lDiff
				bestAdd = guessAdd
				if bestDiff != 0 {
					break passesLoop
				}
			}
			//Below our target
			if diff < 0 {
				//Better than the last time
				mul := ((int32(originalSamples) - diff) << 12) / int32(originalSamples)
				guessAdd = uint32((int32(guessAdd) * mul) >> 12)
				guessAdd++
			} else if diff > 0 {
				mul := ((int32(originalSamples) - diff) << 12) / int32(originalSamples)
				guessAdd = uint32((int32(guessAdd) * mul) >> 12)
				guessAdd--
			}
		}
		c.attackRates[i] = uint32(bestAdd)
	}
	for i := uint8(62); i < 76; i++ {
		//This should provide instant volume maximizing
		c.attackRates[i] = uint32(8) << bitsRateShift
	}
}

func (c *SingleChannel) SupportAllWaveforms(enabled bool) {
	if enabled {
		c.waveFormMask = 0x07
	} else {
		c.waveFormMask = 0x00
	}
}

func (c *SingleChannel) WriteReg(reg uint8, opNum int, val uint8) {
	switch reg {
	case 0x20:
		c.op[opNum].Write20(c, val)
	case 0x40:
		c.op[opNum].Write40(val)
	case 0x60:
		c.op[opNum].Write60(c, val)
	case 0x80:
		c.op[opNum].Write80(c, val)
	case 0xE0:
		c.op[opNum].WriteWaveForm(c, val)
	}
}
