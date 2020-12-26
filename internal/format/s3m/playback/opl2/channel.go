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

type synthMode uint8

const (
	sm2AM = synthMode(iota)
	sm2FM
	sm3AM
	sm3FM
	sm4Start
	sm3FMFM
	sm3AMFM
	sm3FMAM
	sm3AMAM
	sm6Start
	sm2Percussion
	sm3Percussion
)

// Channel is a channel (a combination of Operators)
type Channel struct {
	op           [2]Operator
	synthHandler synthMode
	chanData     uint32   //Frequency/octave and derived values
	old          [2]int32 //Old data for feedback

	feedback uint8 //Feedback shift
	regB0    uint8 //Register values to check for changes
	regC0    uint8
	//This should correspond with reg104, bit 6 indicates a Percussion channel, bit 7 indicates a silent channel
	fourMask  uint8
	maskLeft  int8 //Sign extended values for both channel's panning
	maskRight int8
}

// NewChannel returns a new Channel
func NewChannel() *Channel {
	c := Channel{}
	c.SetupChannel()
	return &c
}

// SetupChannel resets a channel to factory defaults
func (c *Channel) SetupChannel() {
	c.feedback = 31
	c.maskLeft = -1
	c.maskRight = -1
	c.synthHandler = sm2FM
	for i := range c.op {
		c.op[i].SetupOperator()
	}
}

// Op gets the operator at index `index`
func (c *Channel) Op(chip *Chip, index uint) *Operator {
	ch := chip.GetChannelByOffset(c, int(index>>1))
	return &ch.op[index&1]
}

const (
	cShiftKSLBase = 16

	cShiftKeyCode = 24
)

// SetChanData sets the channel data for the channel
func (c *Channel) SetChanData(chip *Chip, data uint32) {
	change := c.chanData ^ data
	c.chanData = data
	c.Op(chip, 0).chanData = data
	c.Op(chip, 1).chanData = data
	//Since a frequency update triggered c, always update frequency
	c.Op(chip, 0).UpdateFrequency()
	c.Op(chip, 1).UpdateFrequency()
	if (change & (0xff << cShiftKSLBase)) != 0 {
		c.Op(chip, 0).UpdateAttenuation()
		c.Op(chip, 1).UpdateAttenuation()
	}
	if (change & (0xff << cShiftKeyCode)) != 0 {
		c.Op(chip, 0).UpdateRates(chip)
		c.Op(chip, 1).UpdateRates(chip)
	}
}

// UpdateFrequency updates the frequency setting
func (c *Channel) UpdateFrequency(chip *Chip, fourOp uint8) {
	//Extrace the frequency bits
	data := c.chanData & 0xffff
	kslBase := cKslTable[data>>6]
	keyCode := (data & 0x1c00) >> 9
	if (chip.reg08 & 0x40) != 0 {
		keyCode |= (data & 0x100) >> 8 /* notesel == 1 */
	} else {
		keyCode |= (data & 0x200) >> 9 /* notesel == 0 */
	}
	//Add the keycode and ksl into the highest bits of chanData
	data |= (keyCode << cShiftKeyCode) | (uint32(kslBase) << cShiftKSLBase)
	c.SetChanData(chip, data)
	if (fourOp & 0x3f) != 0 {
		chip.GetChannelByOffset(c, 1).SetChanData(chip, data)
	}
}

// WriteA0 writes to register 0xA0 for the channel (the lo-byte of the frequency)
func (c *Channel) WriteA0(chip *Chip, val uint8) {
	fourOp := uint8(chip.reg104 & uint8(chip.opl3Active) & c.fourMask)
	//Don't handle writes to silent fourop channels
	if fourOp > 0x80 {
		return
	}
	change := uint32((c.chanData ^ uint32(val)) & 0xff)
	if change != 0 {
		c.chanData ^= change
		c.UpdateFrequency(chip, fourOp)
	}
}

// WriteB0 writes to register 0xB0 for the channel (the hi-byte of the frequency)
func (c *Channel) WriteB0(chip *Chip, val uint8) {
	fourOp := uint8(chip.reg104 & uint8(chip.opl3Active) & c.fourMask)
	//Don't handle writes to silent fourop channels
	if fourOp > 0x80 {
		return
	}
	change := uint((c.chanData ^ (uint32(val) << 8)) & 0x1f00)
	if change != 0 {
		c.chanData ^= uint32(change)
		c.UpdateFrequency(chip, fourOp)
	}
	//Check for a change in the keyon/off state
	if ((val ^ c.regB0) & 0x20) == 0 {
		return
	}
	c.regB0 = val
	if (val & 0x20) != 0 {
		c.Op(chip, 0).KeyOn(0x1)
		c.Op(chip, 1).KeyOn(0x1)
		if (fourOp & 0x3f) != 0 {
			chip.GetChannelByOffset(c, 1).Op(chip, 0).KeyOn(1)
			chip.GetChannelByOffset(c, 1).Op(chip, 1).KeyOn(1)
		}
	} else {
		c.Op(chip, 0).KeyOff(0x1)
		c.Op(chip, 1).KeyOff(0x1)
		if (fourOp & 0x3f) != 0 {
			chip.GetChannelByOffset(c, 1).Op(chip, 0).KeyOff(1)
			chip.GetChannelByOffset(c, 1).Op(chip, 1).KeyOff(1)
		}
	}
}

// GetKeyOn returns true if the Channel's key-on bit is set
func (c *Channel) GetKeyOn() bool {
	return (c.regB0 & 0x20) != 0
}

// WriteC0 writes to register 0xC0 for the channel (the waveform, modulation feedback values, and mode settings)
func (c *Channel) WriteC0(chip *Chip, val uint8) {
	change := val ^ c.regC0
	if change == 0 {
		return
	}
	c.regC0 = val
	c.feedback = (val >> 1) & 7
	if c.feedback != 0 {
		//We shift the input to the right 10 bit wave index value
		c.feedback = 9 - c.feedback
	} else {
		c.feedback = 31
	}
	//Select the new synth mode
	if chip.opl3Active != 0 {
		//4-op mode enabled for c channel
		if ((chip.reg104 & c.fourMask) & 0x3f) != 0 {
			var chan0 *Channel
			var chan1 *Channel
			//Check if it's the 2nd channel in a 4-op
			if (c.fourMask & 0x80) == 0 {
				chan0 = c
				chan1 = chip.GetChannelByOffset(c, 1)
			} else {
				chan0 = chip.GetChannelByOffset(c, -1)
				chan1 = c
			}

			synth := uint8((chan0.regC0&1)<<0) | ((chan1.regC0 & 1) << 1)
			switch synth {
			case 0:
				chan0.synthHandler = sm3FMFM
			case 1:
				chan0.synthHandler = sm3AMFM
			case 2:
				chan0.synthHandler = sm3FMAM
			case 3:
				chan0.synthHandler = sm3AMAM
			}
			//Disable updating percussion channels
		} else if (c.fourMask&0x40) != 0 && (chip.regBD&0x20) != 0 {

			//Regular dual op, am or fm
		} else if (val & 1) != 0 {
			c.synthHandler = sm3AM
		} else {
			c.synthHandler = sm3FM
		}
		if (val & 0x10) != 0 {
			c.maskLeft = -1
		} else {
			c.maskLeft = 0
		}
		if (val & 0x20) != 0 {
			c.maskRight = -1
		} else {
			c.maskRight = 0
		}
		//opl2 active
	} else {
		//Disable updating percussion channels
		if (c.fourMask&0x40) != 0 && (chip.regBD&0x20) != 0 {

			//Regular dual op, am or fm
		} else if (val & 1) != 0 {
			c.synthHandler = sm2AM
		} else {
			c.synthHandler = sm2FM
		}
	}
}

// ResetC0 zorches the register 0xC0
func (c *Channel) ResetC0(chip *Chip) {
	val := uint8(c.regC0)
	c.regC0 ^= 0xff
	c.WriteC0(chip, val)
}

// GeneratePercussion generates percussion data in the channel
func (c *Channel) GeneratePercussion(chip *Chip, output []int32, opl3Mode bool) {
	//BassDrum
	mod := int((c.old[0] + c.old[1]) >> c.feedback)
	c.old[0] = c.old[1]
	c.old[1] = int32(c.Op(chip, 0).GetSample(mod))

	//When bassdrum is in AM mode first operator is ignoed
	if (c.regC0 & 1) != 0 {
		mod = 0
	} else {
		mod = int(c.old[0])
	}
	sample := int32(c.Op(chip, 1).GetSample(mod))

	//Precalculate stuff used by other outputs
	noiseBit := uint32(chip.ForwardNoise() & 0x1)
	c2 := uint32(c.Op(chip, 2).ForwardWave())
	c5 := uint32(c.Op(chip, 5).ForwardWave())
	var phaseBit uint32
	if (((c2 & 0x88) ^ ((c2 << 5) & 0x80)) | ((c5 ^ (c5 << 2)) & 0x20)) != 0 {
		phaseBit = 0x02
	} else {
		phaseBit = 0x00
	}

	//Hi-Hat
	hhVol := c.Op(chip, 2).ForwardVolume()
	if !envSilent(int(hhVol)) {
		hhIndex := uint32((phaseBit << 8) | (0x34 << (phaseBit ^ (noiseBit << 1))))
		sample += int32(c.Op(chip, 2).GetWave(uint(hhIndex), hhVol))
	}
	//Snare Drum
	sdVol := c.Op(chip, 3).ForwardVolume()
	if !envSilent(int(sdVol)) {
		sdIndex := uint32((0x100 + (c2 & 0x100)) ^ (noiseBit << 8))
		sample += int32(c.Op(chip, 3).GetWave(uint(sdIndex), sdVol))
	}
	//Tom-tom
	sample += int32(c.Op(chip, 4).GetSample(0))

	//Top-Cymbal
	tcVol := c.Op(chip, 5).ForwardVolume()
	if !envSilent(int(tcVol)) {
		tcIndex := uint32((1 + phaseBit) << 8)
		sample += int32(c.Op(chip, 5).GetWave(uint(tcIndex), tcVol))
	}
	sample <<= 1
	if opl3Mode {
		output[0] += sample
		output[1] += sample
	} else {
		output[0] += sample
	}
}

// BlockTemplate simulates waveform and envelope data from the channel
func (c *Channel) BlockTemplate(chip *Chip, samples uint32, output []int32, mode synthMode) (int, bool) {
	switch mode {
	case sm2AM, sm3AM:
		if c.Op(chip, 0).Silent() && c.Op(chip, 1).Silent() {
			c.old[0] = 0
			c.old[1] = 0
			return 1, true
		}
	case sm2FM, sm3FM:
		if c.Op(chip, 1).Silent() {
			c.old[0] = 0
			c.old[1] = 0
			return 1, true
		}
	case sm3FMFM:
		if c.Op(chip, 3).Silent() {
			c.old[0] = 0
			c.old[1] = 0
			return 2, true
		}
	case sm3AMFM:
		if c.Op(chip, 0).Silent() && c.Op(chip, 3).Silent() {
			c.old[0] = 0
			c.old[1] = 0
			return 2, true
		}
	case sm3FMAM:
		if c.Op(chip, 1).Silent() && c.Op(chip, 3).Silent() {
			c.old[0] = 0
			c.old[1] = 0
			return 2, true
		}
	case sm3AMAM:
		if c.Op(chip, 0).Silent() && c.Op(chip, 2).Silent() && c.Op(chip, 3).Silent() {
			c.old[0] = 0
			c.old[1] = 0
			return 2, true
		}
	}
	//Init the operators with the the current vibrato and tremolo values
	c.Op(chip, 0).Prepare(chip)
	c.Op(chip, 1).Prepare(chip)
	if mode > sm4Start {
		c.Op(chip, 2).Prepare(chip)
		c.Op(chip, 3).Prepare(chip)
	}
	if mode > sm6Start {
		c.Op(chip, 4).Prepare(chip)
		c.Op(chip, 5).Prepare(chip)
	}
	for i := uint(0); i < uint(samples); i++ {
		//Early out for percussion handlers
		if mode == sm2Percussion {
			c.GeneratePercussion(chip, output[i:], false)
			continue //Prevent some unitialized value bitching
		} else if mode == sm3Percussion {
			c.GeneratePercussion(chip, output[i*2:], true)
			continue //Prevent some unitialized value bitching
		}

		//Do unsigned shift so we can shift out all bits but still stay in 10 bit range otherwise
		mod := int(uint32(c.old[0]+c.old[1]) >> c.feedback)
		c.old[0] = c.old[1]
		c.old[1] = int32(c.Op(chip, 0).GetSample(mod))
		var sample int32
		out0 := int(c.old[0])
		if mode == sm2AM || mode == sm3AM {
			sample = int32(out0 + c.Op(chip, 1).GetSample(0))
		} else if mode == sm2FM || mode == sm3FM {
			sample = int32(c.Op(chip, 1).GetSample(out0))
		} else if mode == sm3FMFM {
			next := int(c.Op(chip, 1).GetSample(out0))
			next = c.Op(chip, 2).GetSample(next)
			sample = int32(c.Op(chip, 3).GetSample(next))
		} else if mode == sm3AMFM {
			sample = int32(out0)
			next := int(c.Op(chip, 1).GetSample(0))
			next = c.Op(chip, 2).GetSample(next)
			sample += int32(c.Op(chip, 3).GetSample(next))
		} else if mode == sm3FMAM {
			sample = int32(c.Op(chip, 1).GetSample(out0))
			next := int(c.Op(chip, 2).GetSample(0))
			sample += int32(c.Op(chip, 3).GetSample(next))
		} else if mode == sm3AMAM {
			sample = int32(out0)
			next := int(c.Op(chip, 1).GetSample(0))
			sample += int32(c.Op(chip, 2).GetSample(next))
			sample += int32(c.Op(chip, 3).GetSample(0))
		}
		switch mode {
		case sm2AM, sm2FM:
			output[i] += sample
		case sm3AM, sm3FM, sm3FMFM, sm3AMFM, sm3FMAM, sm3AMAM:
			output[i*2+0] += sample & int32(c.maskLeft)
			output[i*2+1] += sample & int32(c.maskRight)
		}
	}
	switch mode {
	case sm2AM, sm2FM, sm3AM, sm3FM:
		return 1, true
	case sm3FMFM, sm3AMFM, sm3FMAM, sm3AMAM:
		return 2, true
	case sm2Percussion, sm3Percussion:
		return 3, true
	}
	return 0, false
}
