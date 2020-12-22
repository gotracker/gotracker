package opl2

import "math"

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

	AdditiveSynthesis bool

	chanData uint32   //Frequency/octave and derived values
	old      [2]int32 //Old data for feedback

	feedback uint8 //Feedback shift
	KeyOn    bool
	Block    uint8
	FreqHi   uint8
	regC0    uint8
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

func (c *SingleChannel) WriteA0(val uint8) {
	change := uint32((c.chanData ^ uint32(val)) & 0xff)
	if change != 0 {
		c.chanData ^= change
		c.UpdateFrequency()
	}
}

func (c *SingleChannel) WriteFNum(fnum uint16, block uint8) {
	change := uint((c.chanData ^ (uint32(fnum) | uint32(block)<<10)) & 0x7fff)
	if change != 0 {
		c.chanData ^= uint32(change)
		c.UpdateFrequency()
	}
}

func (c *SingleChannel) SetKeyOn(on bool) {
	if on == c.KeyOn {
		return
	}
	c.KeyOn = on
	if on {
		c.op[0].KeyOn(0x1)
		c.op[1].KeyOn(0x1)
	} else {
		c.op[0].KeyOff(0x1)
		c.op[1].KeyOff(0x1)
	}
}

func (c *SingleChannel) GetKeyOn() bool {
	return c.KeyOn
}

func (c *SingleChannel) WriteC0(val uint8) {
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
	if (val & 1) != 0 {
		c.AdditiveSynthesis = true
	} else {
		c.AdditiveSynthesis = false
	}
}

func (c *SingleChannel) ResetC0() {
	val := uint8(c.regC0)
	c.regC0 ^= 0xff
	c.WriteC0(val)
}

func (c *SingleChannel) BlockTemplate(samples uint32, output []int32) {
	if c.AdditiveSynthesis {
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
		if c.AdditiveSynthesis {
			sample = int32(out0 + c.op[1].GetSample(0))
		} else {
			sample = int32(c.op[1].GetSample(out0))
		}
		output[i] += sample
	}
}

func (c *SingleChannel) ForwardNoise() uint32 {
	c.noiseCounter += c.noiseAdd
	count := uint(c.noiseCounter >> LFO_SH)
	c.noiseCounter &= WAVE_MASK
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
	todo := uint32(LFO_MAX) - c.lfoCounter
	count := uint32((todo + c.lfoAdd - 1) / c.lfoAdd)
	if count > samples {
		count = samples
		c.lfoCounter += count * c.lfoAdd
	} else {
		c.lfoCounter += count * c.lfoAdd
		c.lfoCounter &= uint32(LFO_MAX - 1)
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
	c.AdditiveSynthesis = false
	for i := range c.op {
		c.op[i].SetupOperator()
	}

	//Noise counter is run at the same precision as general waves
	c.noiseAdd = (uint32)(0.5 + scale*float64(uint32(1)<<LFO_SH))
	c.noiseCounter = 0
	c.noiseValue = 1 //Make sure it triggers the noise xor the first time
	//The low frequency oscillation counter
	//Every time his overflows vibrato and tremoloindex are increased
	c.lfoAdd = uint32(0.5 + scale*float64(uint32(1)<<LFO_SH))
	c.lfoCounter = 0
	c.vibratoIndex = 0
	c.tremoloIndex = 0

	//With higher octave this gets shifted up
	//-1 since the freqCreateTable = *2
	if WAVE_PRECISION != 0 {
		freqScale := float64(float64(1<<7) * scale * float64(uint(1<<(WAVE_SH-1-10))))
		for i := 0; i < 16; i++ {
			c.freqMul[i] = uint32(0.5 + freqScale*float64(freqCreateTable[i]))
		}
	} else {
		freqScale := uint32(0.5 + scale*float64(uint(1<<(WAVE_SH-1-10))))
		for i := 0; i < 16; i++ {
			c.freqMul[i] = freqScale * freqCreateTable[i]
		}
	}

	//-3 since the real envelope takes 8 steps to reach the single value we supply
	for i := uint8(0); i < 76; i++ {
		index, shift := EnvelopeSelect(i)
		c.linearRates[i] = uint32(scale * float64(envelopeIncreaseTable[index]<<(RATE_SH+ENV_EXTRA-shift-3)))
	}
	//Generate the best matching attack rate
	for i := uint8(0); i < 62; i++ {
		index, shift := EnvelopeSelect(i)
		//Original amount of samples the attack would take
		original := int32(float64(attackSamplesTable[index]<<shift) / scale)

		guessAdd := int32(scale * float64(envelopeIncreaseTable[index]<<(RATE_SH-shift-3)))
		bestAdd := guessAdd
		bestDiff := uint32(1 << 30)
		for passes := uint32(0); passes < 16; passes++ {
			volume := int32(ENV_MAX)
			samples := int32(0)
			count := uint32(0)
			for volume > 0 && samples < original*2 {
				count += uint32(guessAdd)
				change := int32(count >> RATE_SH)
				count &= RATE_MASK
				if change != 0 { // less than 1 %
					volume += (^volume * change) >> 3
				}
				samples++

			}
			diff := int32(original - samples)
			lDiff := uint32(math.Abs(float64(diff)))
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
				mul := int32(((original - diff) << 12) / original)
				guessAdd = ((guessAdd * mul) >> 12)
				guessAdd++
			} else if diff > 0 {
				mul := int32(((original - diff) << 12) / original)
				guessAdd = (guessAdd * mul) >> 12
				guessAdd--
			}
		}
		c.attackRates[i] = uint32(bestAdd)
	}
	for i := uint8(62); i < 76; i++ {
		//This should provide instant volume maximizing
		c.attackRates[i] = 8 << RATE_SH
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

const (
	WAVE_HANDLER = iota
	WAVE_TABLELOG
	WAVE_TABLEMUL
)

const (
	OPLRATE = 14318180.0 / 288.0

	TREMOLO_TABLE = 52

	//Try to use most precision for frequencies
	//Else try to keep different waves in synch
	//WAVE_PRECISION = 1
	WAVE_PRECISION = 0

	DBOPL_WAVE = WAVE_TABLEMUL

	//Maximum amount of attenuation int
	//Envelope goes to 511, 9 int
	ENV_BITS = 9

	ENV_MIN   = 0
	ENV_EXTRA = ENV_BITS - 9
	ENV_MAX   = 511 << ENV_EXTRA
	ENV_LIMIT = (12 * 256) >> (3 - ENV_EXTRA)

	WAVE_BITS = 10 + (WAVE_PRECISION * 4)
	WAVE_SH   = 32 - WAVE_BITS
	WAVE_MASK = (1 << WAVE_SH) - 1

	//Use the same accuracy as the waves
	LFO_SH = WAVE_SH - 10
	//LFO is controlled by our tremolo 256 sample limit
	LFO_MAX = 256 << LFO_SH

	//Attack/decay/release rate counter shift
	RATE_SH   = 24
	RATE_MASK = (1 << RATE_SH) - 1

	//Has to fit within 16bit lookuptable
	MUL_SH = 16
)

func ENV_SILENT(x int) bool {
	return x >= ENV_LIMIT
}

// Generate the different waveforms out of the sine/exponetial table using handlers
func makeVolume(wave uint, volume uint) int {
	total := wave + volume
	index := total & 0xff
	sig := uint(expTable[index])
	exp := total >> 8
	return int(sig >> exp)
}

func waveForm0(i uint, volume uint) int {
	neg := int(0 - ((i >> 9) & 1)) //Create ~0 or 0
	wave := uint(sinTable[i&511])
	return (makeVolume(wave, volume) ^ neg) - neg
}

func waveForm1(i uint, volume uint) int {
	wave := uint(sinTable[i&511])
	wave |= (((i ^ 512) & 512) - 1) >> (32 - 12)
	return makeVolume(wave, volume)
}

func waveForm2(i uint, volume uint) int {
	wave := uint(sinTable[i&511])
	return makeVolume(wave, volume)
}

func waveForm3(i uint, volume uint) int {
	wave := uint(sinTable[i&255])
	wave |= (((i ^ 256) & 256) - 1) >> (32 - 12)
	return makeVolume(wave, volume)
}

func waveForm4(i uint, volume uint) int {
	//Twice as fast
	i <<= 1
	neg := int(0 - ((i >> 9) & 1)) //Create ~0 or 0
	wave := uint(sinTable[i&511])
	wave |= (((i ^ 512) & 512) - 1) >> (32 - 12)
	return (makeVolume(wave, volume) ^ neg) - neg
}

func waveForm5(i uint, volume uint) int {
	//Twice as fast
	i <<= 1
	wave := uint(sinTable[i&511])
	wave |= (((i ^ 512) & 512) - 1) >> (32 - 12)
	return makeVolume(wave, volume)
}
func waveForm6(i uint, volume uint) int {
	neg := int(0 - ((i >> 9) & 1)) //Create ~0 or 0
	return (makeVolume(0, volume) ^ neg) - neg
}
func waveForm7(i uint, volume uint) int {
	//Negative is reversed here
	neg := int(((i >> 9) & 1) - 1)
	wave := (i << 3)
	//When negative the volume also runs backwards
	wave = uint(((int(wave) ^ neg) - neg) & 4095)
	return (makeVolume(wave, volume) ^ neg) - neg
}

var waveHandlerTable = [8]waveHandler{
	waveForm0, waveForm1, waveForm2, waveForm3,
	waveForm4, waveForm5, waveForm6, waveForm7,
}

//The lower bits are the shift of the operator vibrato value
//The highest bit is right shifted to generate -1 or 0 for negation
//So taking the highest input value of 7 this gives 3, 7, 3, 0, -3, -7, -3, 0
var vibratoTable = [8]int8{
	1 - 0x00, 0 - 0x00, 1 - 0x00, 30 - 0x00,
	1 - 0x80, 0 - 0x80, 1 - 0x80, 30 - 0x80,
}

//How much to substract from the base value for the final attenuation
var kslCreateTable = [16]uint8{
	//0 will always be be lower than 7 * 8
	64, 32, 24, 19,
	16, 12, 11, 10,
	8, 6, 5, 4,
	3, 2, 1, 0,
}

func m1(x float64) uint32 {
	return uint32(x * 2)
}

var freqCreateTable = [16]uint32{
	m1(0.5), m1(1), m1(2), m1(3), m1(4), m1(5), m1(6), m1(7),
	m1(8), m1(9), m1(10), m1(10), m1(12), m1(12), m1(15), m1(15),
}

//We're not including the highest attack rate, that gets a special value
var attackSamplesTable = [13]uint8{
	69, 55, 46, 40,
	35, 29, 23, 20,
	19, 15, 11, 10,
	9,
}

//On a real opl these values take 8 samples to reach and are based upon larger tables
var envelopeIncreaseTable = [13]uint8{
	4, 5, 6, 7,
	8, 10, 12, 14,
	16, 20, 24, 28,
	32,
}

var expTable = make([]uint16, 256)

//PI table used by WAVEHANDLER
var sinTable = make([]uint16, 512)

//Layout of the waveform table in 512 entry intervals
//With overlapping waves we reduce the table to half it's size

//	|    |//\\|____|WAV7|//__|/\  |____|/\/\|
//	|\\//|    |    |WAV7|    |  \/|    |    |
//	|06  |0126|17  |7   |3   |4   |4 5 |5   |

//6 is just 0 shifted and masked

var waveTable = make([]int16, 8*512)

//Distance into WaveTable the wave starts
var waveBaseTable = [8]uint16{
	0x000, 0x200, 0x200, 0x800,
	0xa00, 0xc00, 0x100, 0x400,
}

//Mask the counter with this
var waveMaskTable = [8]uint16{
	1023, 1023, 511, 511,
	1023, 1023, 512, 1023,
}

//Where to start the counter on at keyon
var waveStartTable = [8]uint16{
	512, 0, 0, 0,
	0, 512, 512, 256,
}

var mulTable = make([]uint16, 384)

var kslTable = make([]uint8, 8*16)
var tremoloTable = make([]uint8, TREMOLO_TABLE)

//Generate a table index and table shift value using input value from a selected rate
func EnvelopeSelect(val uint8) (index uint8, shift uint8) {
	if val < 13*4 { //Rate 0 - 12
		shift = 12 - (val >> 2)
		index = val & 3
	} else if val < 15*4 { //rate 13 - 14
		shift = 0
		index = val - 12*4
	} else { //rate 15 and up
		shift = 0
		index = 12
	}
	return
}

func init() {
	if DBOPL_WAVE == WAVE_HANDLER || DBOPL_WAVE == WAVE_TABLELOG {
		//Exponential volume table, same as the real adlib
		for i := 0; i < 256; i++ {
			//Save them in reverse
			expTable[i] = uint16(0.5 + (math.Pow(2.0, float64(255-i)*(1.0/256))-1)*1024)
			expTable[i] += 1024 //or remove the -1 oh well :)
			//Preshift to the left once so the final volume can shift to the right
			expTable[i] *= 2
		}
	}

	if DBOPL_WAVE == WAVE_HANDLER {
		//Add 0.5 for the trunc rounding of the integer cast
		//Do a PI sinetable instead of the original 0.5 PI
		for i := 0; i < 512; i++ {
			sinTable[i] = uint16(int16((0.5 - math.Log10(math.Sin((float64(i)+0.5)*(math.Pi/512.0)))/math.Log10(2.0)*256)))
		}
	}

	if DBOPL_WAVE == WAVE_TABLEMUL {
		//Multiplication based tables
		for i := 0; i < 384; i++ {
			s := int(i * 8)
			//TODO maybe keep some of the precision errors of the original table?
			val := float64((0.5 + (math.Pow(2.0, -1.0+float64(255-s)*(1.0/256)))*(1<<MUL_SH)))
			mulTable[i] = uint16(val)
		}

		//Sine Wave Base
		for i := 0; i < 512; i++ {
			waveTable[0x0200+i] = int16((math.Sin((float64(i)+0.5)*(math.Pi/512.0)) * 4084))
			waveTable[0x0000+i] = -waveTable[0x200+i]
		}
		//Exponential wave
		for i := 0; i < 256; i++ {
			waveTable[0x700+i] = int16((0.5 + (math.Pow(2.0, -1.0+float64(255-i*8)*(1.0/256)))*4085))
			waveTable[0x6ff-i] = -waveTable[0x700+i]
		}
	}

	if DBOPL_WAVE == WAVE_TABLELOG {
		//Sine Wave Base
		for i := 0; i < 512; i++ {
			waveTable[0x0200+i] = int16((0.5 - math.Log10(math.Sin((float64(i)+0.5)*(math.Pi/512.0)))/math.Log10(2.0)*256))
			waveTable[0x0000+i] = int16((uint16(0x8000) | uint16(waveTable[0x200+i])))
		}
		//Exponential wave
		for i := 0; i < 256; i++ {
			waveTable[0x700+i] = int16(i * 8)
			waveTable[0x6ff-i] = int16(0x8000 | i*8)
		}
	}

	//	|    |//\\|____|WAV7|//__|/\  |____|/\/\|
	//	|\\//|    |    |WAV7|    |  \/|    |    |
	//	|06  |0126|27  |7   |3   |4   |4 5 |5   |

	if DBOPL_WAVE == WAVE_TABLELOG || DBOPL_WAVE == WAVE_TABLEMUL {
		for i := 0; i < 256; i++ {
			//Fill silence gaps
			waveTable[0x400+i] = waveTable[0]
			waveTable[0x500+i] = waveTable[0]
			waveTable[0x900+i] = waveTable[0]
			waveTable[0xc00+i] = waveTable[0]
			waveTable[0xd00+i] = waveTable[0]
			//Replicate sines in other pieces
			waveTable[0x800+i] = waveTable[0x200+i]
			//float64 speed sines
			waveTable[0xa00+i] = waveTable[0x200+i*2]
			waveTable[0xb00+i] = waveTable[0x000+i*2]
			waveTable[0xe00+i] = waveTable[0x200+i*2]
			waveTable[0xf00+i] = waveTable[0x200+i*2]
		}
	}

	//Create the ksl table
	for oct := int(0); oct < 8; oct++ {
		base := int(oct * 8)
		for i := 0; i < 16; i++ {
			val := base - int(kslCreateTable[i])
			if val < 0 {
				val = 0
			}
			// *4 for the final range to match attenuation range
			kslTable[oct*16+i] = uint8(val * 4)
		}
	}
	//Create the Tremolo table, just increase and decrease a triangle wave
	for i := uint8(0); i < TREMOLO_TABLE/2; i++ {
		val := uint8(i << ENV_EXTRA)
		tremoloTable[i] = val
		tremoloTable[TREMOLO_TABLE-1-i] = val
	}
}
