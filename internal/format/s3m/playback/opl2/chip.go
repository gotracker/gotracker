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

// Chip is the current state and emulator of the YM3812/YM262 OPL2/3 chip
type Chip struct {
	//This is used as the base counter for vibrato and tremolo
	lfoCounter uint32
	lfoAdd     uint32

	noiseCounter uint32
	noiseAdd     uint32
	noiseValue   uint32

	//Frequency scales for the different multiplications
	freqMul [16]uint32
	//Rates for decay and release for rate of this chip
	linearRates [76]uint32
	//Best match attack rates for the rate of this chip
	attackRates [76]uint32

	//18 channels with 2 operators each
	ch [18]Channel

	reg104          uint8
	reg08           uint8
	reg04           uint8
	regBD           uint8
	vibratoIndex    uint8
	tremoloIndex    uint8
	vibratoSign     int8
	vibratoShift    uint8
	tremoloValue    uint8
	vibratoStrength uint8
	tremoloStrength uint8
	//Mask for allowed wave forms
	waveFormMask uint8
	//0 or -1 when enabled
	opl3Active int8

	isOPL3 int
}

// NewChip creates a new Chip object
func NewChip(rate uint32, isOPL3 bool) *Chip {
	c := &Chip{}
	for i := range c.ch {
		c.ch[i].SetupChannel()
	}

	var chipIsOPL3 int
	if isOPL3 {
		chipIsOPL3 = -1
	} else {
		chipIsOPL3 = 0
	}
	c.Setup(rate, chipIsOPL3)
	return c
}

// GetChannelByOffset returns the channel `ofs` units away from the `ch` channel
func (c *Chip) GetChannelByOffset(ch *Channel, ofs int) *Channel {
	ci := c.GetChannelIndex(ch)
	if ci < 0 {
		return nil
	}
	return c.GetChannelByIndex(uint32(ci + ofs))
}

// GetChannelIndex gets the channel index (with skips in-built) for `ch`
func (c *Chip) GetChannelIndex(ch *Channel) int {
	for i := uint32(0); i < 32; i++ {
		cc := c.GetChannelByIndex(i)
		if cc == ch {
			return int(i)
		}
	}
	return -1
}

// GetChannelByIndex gets the channel at (skip-laiden) index `i`
func (c *Chip) GetChannelByIndex(i uint32) *Channel {
	index := i & 0xf
	if index >= 9 {
		return nil
	}
	//Make sure the four op channels follow eachother
	if index < 6 {
		index = (index%3)*2 + (index / 3)
	}
	//Add back the bits for highest ones
	if i >= 16 {
		index += 9
	}
	return &c.ch[index]
}

// GetOperatorByIndex gets the operator indexed by `i` that satisfies the channel offset gap (skips)
func (c *Chip) GetOperatorByIndex(i uint32) *Operator {
	if i%8 >= 6 || (i/8)%4 == 3 {
		return nil
	}
	chNum := (i/8)*3 + (i%8)%3
	//Make sure we use 16 and up for the 2nd range to match the chanoffset gap
	if chNum >= 12 {
		chNum += 16 - 12
	}
	opNum := (i % 8) / 3
	if int(chNum) < len(c.ch) {
		return &c.ch[chNum].op[opNum]
	}
	return nil
}

// ForwardNoise updates the noise values and returns the new value
func (c *Chip) ForwardNoise() uint32 {
	c.noiseCounter += c.noiseAdd
	count := uint(c.noiseCounter) >> cLFOSh
	c.noiseCounter &= cWaveMask
	for ; count > 0; count-- {
		//Noise calculation from mame
		c.noiseValue ^= (0x800302) & (0 - (c.noiseValue & 1))
		c.noiseValue >>= 1
	}
	return c.noiseValue
}

// ForwardLFO updates the internal LFOs and returns the amount of samples they updated by
func (c *Chip) ForwardLFO(samples uint32) uint32 {
	//Current vibrato value, runs 4x slower than tremolo
	vibVal := cVibratoTable[c.vibratoIndex>>2]
	c.vibratoSign = 0
	if vibVal < 0 {
		c.vibratoSign = -1
	}
	c.vibratoShift = uint8(vibVal)&7 + c.vibratoStrength
	c.tremoloValue = cTremoloTable[c.tremoloIndex] >> c.tremoloStrength

	//Check hom many samples there can be done before the value changes
	todo := uint32(cLFOMax) - c.lfoCounter
	count := (todo + c.lfoAdd - 1) / c.lfoAdd
	if count > samples {
		count = samples
		c.lfoCounter += count * c.lfoAdd
	} else {
		c.lfoCounter += count * c.lfoAdd
		c.lfoCounter &= uint32(cLFOMax) - 1
		//Maximum of 7 vibrato value * 4
		c.vibratoIndex = (c.vibratoIndex + 1) & 31
		//Clip tremolo to the the table size
		if c.tremoloIndex+1 < cTremoloTableSize {
			c.tremoloIndex++
		} else {
			c.tremoloIndex = 0
		}
	}
	return count
}

// WriteBD writes directly to register 0xBD
func (c *Chip) WriteBD(val uint8) {
	change := c.regBD ^ val
	if change == 0 {
		return
	}
	c.regBD = val
	//TODO could do this with shift and xor?
	if (val & 0x40) != 0 {
		c.vibratoStrength = 0x00
	} else {
		c.vibratoStrength = 0x01
	}
	if (val & 0x80) != 0 {
		c.tremoloStrength = 0x00
	} else {
		c.tremoloStrength = 0x02
	}
	if (val & 0x20) != 0 {
		//Drum was just enabled, make sure channel 6 has the right synth
		if (change & 0x20) != 0 {
			if c.opl3Active != 0 {
				c.ch[6].synthHandler = sm3Percussion
			} else {
				c.ch[6].synthHandler = sm2Percussion
			}
		}
		//Bass Drum
		if (val & 0x10) != 0 {
			c.ch[6].op[0].KeyOn(0x2)
			c.ch[6].op[1].KeyOn(0x2)
		} else {
			c.ch[6].op[0].KeyOff(0x2)
			c.ch[6].op[1].KeyOff(0x2)
		}
		//Hi-Hat
		if (val & 0x1) != 0 {
			c.ch[7].op[0].KeyOn(0x2)
		} else {
			c.ch[7].op[0].KeyOff(0x2)
		}
		//Snare
		if (val & 0x8) != 0 {
			c.ch[7].op[1].KeyOn(0x2)
		} else {
			c.ch[7].op[1].KeyOff(0x2)
		}
		//Tom-Tom
		if (val & 0x4) != 0 {
			c.ch[8].op[0].KeyOn(0x2)
		} else {
			c.ch[8].op[0].KeyOff(0x2)
		}
		//Top Cymbal
		if (val & 0x2) != 0 {
			c.ch[8].op[1].KeyOn(0x2)
		} else {
			c.ch[8].op[1].KeyOff(0x2)
		}
		//Toggle keyoffs when we turn off the percussion
	} else if (change & 0x20) != 0 {
		//Trigger a reset to setup the original synth handler
		c.ch[6].ResetC0(c)
		c.ch[6].op[0].KeyOff(0x2)
		c.ch[6].op[1].KeyOff(0x2)
		c.ch[7].op[0].KeyOff(0x2)
		c.ch[7].op[1].KeyOff(0x2)
		c.ch[8].op[0].KeyOff(0x2)
		c.ch[8].op[1].KeyOff(0x2)
	}
}

// WriteReg writes to register `reg` with value `val`
func (c *Chip) WriteReg(reg uint32, val uint8) {
	switch (reg & 0xf0) >> 4 {
	case 0x00 >> 4:
		if reg == 0x01 {
			if (val & 0x20) != 0 {
				c.waveFormMask = 0x7
			} else {
				c.waveFormMask = 0x0
			}
		} else if reg == 0x104 {
			//Only detect changes in lowest 6 bits
			if ((c.reg104 ^ val) & 0x3f) == 0 {
				return
			}
			//Always keep the highest bit enabled, for checking > 0x80
			c.reg104 = 0x80 | (val & 0x3f)
		} else if reg == 0x105 {
			//MAME says the real opl3 doesn't reset anything on opl3 disable/enable till the next write in another register
			if ((uint8(c.opl3Active) ^ val) & 1) == 0 {
				return
			}
			if (val & 1) != 0 {
				c.opl3Active = -1
			} else {
				c.opl3Active = 0
			}
			//Update the 0xc0 register for all channels to signal the switch to mono/stereo handlers
			for i := 0; i < 18; i++ {
				c.ch[i].ResetC0(c)
			}
		} else if reg == 0x08 {
			c.reg08 = val
		}
	case 0x10 >> 4:
	case 0x20 >> 4, 0x30 >> 4:
		index := ((reg >> 3) & 0x20) | (reg & 0x1f)
		o := c.GetOperatorByIndex(index)
		if o != nil {
			o.Write20(c, val)
		}
	case 0x40 >> 4, 0x50 >> 4:
		index := ((reg >> 3) & 0x20) | (reg & 0x1f)
		o := c.GetOperatorByIndex(index)
		if o != nil {
			o.Write40(c, val)
		}
	case 0x60 >> 4, 0x70 >> 4:
		index := ((reg >> 3) & 0x20) | (reg & 0x1f)
		o := c.GetOperatorByIndex(index)
		if o != nil {
			o.Write60(c, val)
		}
	case 0x80 >> 4, 0x90 >> 4:
		index := ((reg >> 3) & 0x20) | (reg & 0x1f)
		o := c.GetOperatorByIndex(index)
		if o != nil {
			o.Write80(c, val)
		}
	case 0xa0 >> 4:
		index := ((reg >> 4) & 0x10) | (reg & 0xf)
		ch := c.GetChannelByIndex(index)
		if ch != nil {
			ch.WriteA0(c, val)
		}
	case 0xb0 >> 4:
		if reg == 0xbd {
			c.WriteBD(val)
		} else {
			index := ((reg >> 4) & 0x10) | (reg & 0xf)
			ch := c.GetChannelByIndex(index)
			if ch != nil {
				ch.WriteB0(c, val)
			}
		}
	case 0xc0 >> 4:
		index := ((reg >> 4) & 0x10) | (reg & 0xf)
		ch := c.GetChannelByIndex(index)
		if ch != nil {
			ch.WriteC0(c, val)
		}
	case 0xd0 >> 4:
	case 0xe0 >> 4, 0xf0 >> 4:
		index := ((reg >> 3) & 0x20) | (reg & 0x1f)
		o := c.GetOperatorByIndex(index)
		if o != nil {
			o.WriteE0(c, val)
		}
	}
}

// WriteAddr calculates the actual value to be written at a specific port
func (c *Chip) WriteAddr(port uint32, val uint8) uint32 {
	switch port & 3 {
	case 0:
		return uint32(val)
	case 2:
		if c.opl3Active != 0 || val == 0x05 {
			return 0x100 | uint32(val)
		}
		return uint32(val)
	}
	return 0
}

// GenerateBlock2 returns sample data for OPL2 output
func (c *Chip) GenerateBlock2(total uint, output []int32) {
	outputIdx := uint(0)
	for total > 0 {
		samples := c.ForwardLFO(uint32(total))
		count := 0
		for i := 0; i < 9; {
			ch := &c.ch[i]
			count++
			ofs, valid := ch.BlockTemplate(c, samples, output[outputIdx:], ch.synthHandler)
			if !valid {
				panic("invalid offset returned from BlockTemplate")
			}
			i += ofs
		}
		total -= uint(samples)
		outputIdx += uint(samples)
	}
}

// GenerateBlock3 returns sample data for OPL3 output (stereo!)
func (c *Chip) GenerateBlock3(total uint, output []int32) {
	outputIdx := uint(0)
	for total > 0 {
		samples := c.ForwardLFO(uint32(total))
		count := 0
		for i := 0; i < 18; {
			ch := &c.ch[i]
			count++
			ofs, valid := ch.BlockTemplate(c, samples, output[outputIdx:], ch.synthHandler)
			if !valid {
				panic("invalid offset returned from BlockTemplate")
			}
			i += ofs
		}
		total -= uint(samples)
		outputIdx += uint(samples) * 2
	}
}

// Setup sets up a chip for correct operation
func (c *Chip) Setup(rate uint32, chipIsOPL3 int) {
	original := float64(OPLRATE)
	scale := original / float64(rate)

	c.isOPL3 = chipIsOPL3

	//Noise counter is run at the same precision as general waves
	c.noiseAdd = uint32(0.5 + scale*float64(uint32(1)<<cLFOSh))
	c.noiseCounter = 0
	c.noiseValue = 1 //Make sure it triggers the noise xor the first time
	//The low frequency oscillation counter
	//Every time his overflows vibrato and tremoloindex are increased
	c.lfoAdd = uint32(0.5 + scale*float64(uint32(1)<<cLFOSh))
	c.lfoCounter = 0
	c.vibratoIndex = 0
	c.tremoloIndex = 0

	//With higher octave this gets shifted up
	//-1 since the freqCreateTable = *2
	if cWavePrecision != 0 {
		freqScale := float64(float64(1<<7) * scale * float64(uint(1)<<(cWaveSh-1-10)))
		for i := 0; i < 16; i++ {
			c.freqMul[i] = uint32(0.5 + freqScale*float64(cFreqCreateTable[i]))
		}
	} else {
		freqScale := uint32(0.5 + scale*float64(uint(1)<<(cWaveSh-1-10)))
		for i := 0; i < 16; i++ {
			c.freqMul[i] = freqScale * cFreqCreateTable[i]
		}
	}

	//-3 since the real envelope takes 8 steps to reach the single value we supply
	for i := uint8(0); i < 76; i++ {
		index, shift := envelopeSelect(i)
		c.linearRates[i] = uint32(scale * float64(uint(cEnvelopeIncreaseTable[index])<<(cRateSh+cEnvExtra-shift-3)))
	}
	//Generate the best matching attack rate
	for i := uint8(0); i < 62; i++ {
		index, shift := envelopeSelect(i)
		//Original amount of samples the attack would take
		original := int32(float64(uint(cAttackSamplesTable[index])<<shift) / scale)

		guessAdd := int32(scale * float64(uint(cEnvelopeIncreaseTable[index])<<(cRateSh-shift-3)))
		bestAdd := guessAdd
		bestDiff := uint32(1) << 30
		for passes := uint32(0); passes < 16; passes++ {
			volume := int32(cEnvMax)
			samples := int32(0)
			count := uint32(0)
			for volume > 0 && samples < original*2 {
				count += uint32(guessAdd)
				change := int32(count) >> cRateSh
				count &= cRateMask
				if change != 0 { // less than 1 %
					volume += (^volume * change) >> 3
				}
				samples++

			}
			diff := original - samples
			lDiff := uint32(diff)
			if diff < 0 {
				lDiff = uint32(-diff)
			}
			//Init last on first pass
			if lDiff < bestDiff {
				bestDiff := lDiff
				bestAdd = guessAdd
				if bestDiff != 0 {
					break
				}
			}
			//Below our target
			if diff < 0 {
				//Better than the last time
				mul := ((original - diff) << 12) / original
				guessAdd = (guessAdd * mul) >> 12
				guessAdd++
			} else if diff > 0 {
				mul := ((original - diff) << 12) / original
				guessAdd = (guessAdd * mul) >> 12
				guessAdd--
			}
		}
		c.attackRates[i] = uint32(bestAdd)
	}
	for i := uint8(62); i < 76; i++ {
		//This should provide instant volume maximizing
		c.attackRates[i] = uint32(8) << cRateSh
	}
	//Setup the channels with the correct four op flags
	//Channels are accessed through a table so they appear linear here
	c.ch[0].fourMask = 0x00 | (1 << 0)
	c.ch[1].fourMask = 0x80 | (1 << 0)
	c.ch[2].fourMask = 0x00 | (1 << 1)
	c.ch[3].fourMask = 0x80 | (1 << 1)
	c.ch[4].fourMask = 0x00 | (1 << 2)
	c.ch[5].fourMask = 0x80 | (1 << 2)

	c.ch[9].fourMask = 0x00 | (1 << 3)
	c.ch[10].fourMask = 0x80 | (1 << 3)
	c.ch[11].fourMask = 0x00 | (1 << 4)
	c.ch[12].fourMask = 0x80 | (1 << 4)
	c.ch[13].fourMask = 0x00 | (1 << 5)
	c.ch[14].fourMask = 0x80 | (1 << 5)

	//mark the percussion channels
	c.ch[6].fourMask = 0x40
	c.ch[7].fourMask = 0x40
	c.ch[8].fourMask = 0x40

	//Clear Everything in opl3 mode
	c.WriteReg(0x105, 0x1)
	for i := uint32(0); i < 512; i++ {
		if i == 0x105 {
			continue
		}
		c.WriteReg(i, 0xff)
		c.WriteReg(i, 0x0)
	}
	c.WriteReg(0x105, 0x0)
	//Clear everything in opl2 mode
	for i := uint32(0); i < 255; i++ {
		c.WriteReg(i, 0xff)
		c.WriteReg(i, 0x0)
	}
}
