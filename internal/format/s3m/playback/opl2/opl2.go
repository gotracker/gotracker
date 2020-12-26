package opl2

import "math"

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

const (
	cWaveHandler = iota
	cWaveTableLog
	cWaveTableMul
)

const (
	// OPLRATE is the sampling rate that the OPL2/3 outputs samples at, normally
	// all internal calculations are defined by it.
	OPLRATE = 14318180.0 / 288.0

	cTremoloTableSize = 52

	//Try to use most precision for frequencies
	//Else try to keep different waves in synch
	//cWavePrecision = 1
	cWavePrecision = 0

	cDBOPLWave = cWaveTableMul

	//cWavePrecision = 1:
	//  Need some extra bits at the top to have room for octaves and frequency multiplier
	//  We support to 8 times lower rate
	//  128 * 15 * 8 = 15350, 2^13.9, so need 14 bits
	//cWavePrecision = 0:
	//  Wave bits available in the top of the 32bit range
	//  Original adlib uses 10.10, we use 10.22
	cWaveBits = 10 + int(cWavePrecision)*4
	cWaveSh   = 32 - cWaveBits
	cWaveMask = (1 << cWaveSh) - 1

	//Use the same accuracy as the waves
	cLFOSh = cWaveSh - 10
	//LFO is controlled by our tremolo 256 sample limit
	cLFOMax = 256 << cLFOSh

	//Maximum amount of attenuation bits
	//Envelope goes to 511, 9 bits
	cEnvBits = 9

	cEnvMin   = 0
	cEnvExtra = cEnvBits - 9
	cEnvMax   = 511 << cEnvExtra
	cEnvLimit = (12 * 256) >> (3 - cEnvExtra)
)

func envSilent(x int) bool {
	return x >= cEnvLimit
}

const (
	//Attack/decay/release rate counter shift
	cRateSh   = 24
	cRateMask = (1 << cRateSh) - 1
	//Has to fit within 16bit lookuptable
	cMulSh = 16
)

func init() {
	//Check some ranges
	if cEnvExtra > 3 {
		panic("Too many envelope bits")
	}
}

//How much to substract from the base value for the final attenuation
var cKslCreateTable = [16]uint8{
	//0 will always be be lower than 7 * 8
	64, 32, 24, 19,
	16, 12, 11, 10,
	8, 6, 5, 4,
	3, 2, 1, 0,
}

func m1(x float64) uint32 {
	return uint32(x * 2)
}

var cFreqCreateTable = [16]uint32{
	m1(0.5), m1(1), m1(2), m1(3), m1(4), m1(5), m1(6), m1(7),
	m1(8), m1(9), m1(10), m1(10), m1(12), m1(12), m1(15), m1(15),
}

//We're not including the highest attack rate, that gets a special value
var cAttackSamplesTable = [13]uint8{
	69, 55, 46, 40,
	35, 29, 23, 20,
	19, 15, 11, 10,
	9,
}

//On a real opl these values take 8 samples to reach and are based upon larger tables
var cEnvelopeIncreaseTable = [13]uint8{
	4, 5, 6, 7,
	8, 10, 12, 14,
	16, 20, 24, 28,
	32,
}

var cExpTable = make([]uint16, 256)

//PI table used by WAVEHANDLER
var cSinTable = make([]uint16, 512)

//Layout of the waveform table in 512 entry intervals
//With overlapping waves we reduce the table to half it's size

//	|    |//\\|____|WAV7|//__|/\  |____|/\/\|
//	|\\//|    |    |WAV7|    |  \/|    |    |
//	|06  |0126|17  |7   |3   |4   |4 5 |5   |

//6 is just 0 shifted and masked

var cWaveTable = make([]int16, 8*512)

//Distance into WaveTable the wave starts
var cWaveBaseTable = [8]uint16{
	0x000, 0x200, 0x200, 0x800,
	0xa00, 0xc00, 0x100, 0x400,
}

//Mask the counter with this
var cWaveMaskTable = [8]uint16{
	1023, 1023, 511, 511,
	1023, 1023, 512, 1023,
}

//Where to start the counter on at keyon
var cWaveStartTable = [8]uint16{
	512, 0, 0, 0,
	0, 512, 512, 256,
}

var cMulTable = make([]uint16, 384)

var cKslTable = make([]uint8, 8*16)
var cTremoloTable = make([]uint8, cTremoloTableSize)

//Start of a channel behind the chip struct start
var cChanOffsetTable = make([]uint16, 32)

//The lower bits are the shift of the operator vibrato value
//The highest bit is right shifted to generate -1 or 0 for negation
//So taking the highest input value of 7 this gives 3, 7, 3, 0, -3, -7, -3, 0
var cVibratoTable = [8]int8{
	1 - 0x00, 0 - 0x00, 1 - 0x00, 30 - 0x00,
	1 - 0x80, 0 - 0x80, 1 - 0x80, 30 - 0x80,
}

//Shift strength for the ksl value determined by ksl strength
var cKslShiftTable = [4]uint8{
	31, 1, 2, 0,
}

//Generate a table index and table shift value using input value from a selected rate
func envelopeSelect(val uint8) (index uint8, shift uint8) {
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

/*
	Generate the different waveforms out of the sine/exponetial table using handlers
*/
func makeVolume(wave int, volume int) int {
	total := wave + volume
	index := total & 0xff
	sig := uint(cExpTable[index])
	exp := total >> 8
	return int(sig) >> exp
}

func waveForm0(i uint, volume int) int {
	neg := int(0)
	if ((i >> 9) & 1) != 0 {
		neg = -1
	}
	wave := int(cSinTable[i&511])
	oVol := makeVolume(wave, volume)
	vol := oVol ^ neg
	vol -= neg
	return vol
}

func waveForm1(i uint, volume int) int {
	wave := int(cSinTable[i&511])
	wave |= (((int(i) ^ 512) & 512) - 1) >> (32 - 12)
	return makeVolume(wave, volume)
}

func waveForm2(i uint, volume int) int {
	wave := int(cSinTable[i&511])
	return makeVolume(wave, volume)
}

func waveForm3(i uint, volume int) int {
	wave := int(cSinTable[i&255])
	wave |= (((int(i) ^ 256) & 256) - 1) >> (32 - 12)
	return makeVolume(wave, volume)
}

func waveForm4(i uint, volume int) int {
	//Twice as fast
	i <<= 1
	neg := int(0 - ((i >> 9) & 1)) //Create ~0 or 0
	wave := int(cSinTable[i&511])
	wave |= (((int(i) ^ 512) & 512) - 1) >> (32 - 12)
	return (makeVolume(wave, volume) ^ neg) - neg
}

func waveForm5(i uint, volume int) int {
	//Twice as fast
	i <<= 1
	wave := int(cSinTable[i&511])
	wave |= (((int(i) ^ 512) & 512) - 1) >> (32 - 12)
	return makeVolume(wave, volume)
}
func waveForm6(i uint, volume int) int {
	neg := int(0 - ((i >> 9) & 1)) //Create ~0 or 0
	return (makeVolume(0, volume) ^ neg) - neg
}
func waveForm7(i uint, volume int) int {
	//Negative is reversed here
	neg := int(((i >> 9) & 1) - 1)
	wave := int(i) << 3
	//When negative the volume also runs backwards
	wave = ((int(wave) ^ neg) - neg) & 4095
	return (makeVolume(wave, volume) ^ neg) - neg
}

type waveHandler func(uint, int) int

var waveHandlerTable = [8]waveHandler{
	waveForm0, waveForm1, waveForm2, waveForm3,
	waveForm4, waveForm5, waveForm6, waveForm7,
}

func init() {
	if cDBOPLWave == cWaveHandler || cDBOPLWave == cWaveTableLog {
		//Exponential volume table, same as the real adlib
		for i := 0; i < 256; i++ {
			//Save them in reverse
			exp := float64(255-i) / 256.0
			p := math.Pow(2.0, exp) - 1
			expVal := uint16(math.Round(p * 1024))
			expVal += 1024 //or remove the -1 oh well :)
			//Preshift to the left once so the final volume can shift to the right
			cExpTable[i] = expVal * 2
			//ExpTable[i] *= 2
		}
	}

	if cDBOPLWave == cWaveHandler {
		//Add 0.5 for the trunc rounding of the integer cast
		//Do a PI sinetable instead of the original 0.5 PI
		piPiece := math.Pi / 512.0
		for i := 0; i < 512; i++ {
			a := 0.5 - math.Log2(math.Sin((float64(i)+0.5)*piPiece))*256
			cSinTable[i] = uint16(a)
		}
	}

	if cDBOPLWave == cWaveTableMul {
		//Multiplication based tables
		for i := 0; i < 384; i++ {
			s := int(i * 8)
			//TODO maybe keep some of the precision errors of the original table?
			val := float64((0.5 + (math.Pow(2.0, -1.0+float64(255-s)*(1.0/256)))*(1<<cMulSh)))
			cMulTable[i] = uint16(val)
		}

		//Sine Wave Base
		for i := 0; i < 512; i++ {
			cWaveTable[0x0200+i] = int16((math.Sin((float64(i)+0.5)*(math.Pi/512.0)) * 4084))
			cWaveTable[0x0000+i] = -cWaveTable[0x200+i]
		}
		//Exponential wave
		for i := 0; i < 256; i++ {
			cWaveTable[0x700+i] = int16((0.5 + (math.Pow(2.0, -1.0+float64(255-i*8)*(1.0/256)))*4085))
			cWaveTable[0x6ff-i] = -cWaveTable[0x700+i]
		}
	}

	if cDBOPLWave == cWaveTableLog {
		//Sine Wave Base
		for i := 0; i < 512; i++ {
			cWaveTable[0x0200+i] = int16((0.5 - math.Log10(math.Sin((float64(i)+0.5)*(math.Pi/512.0)))/math.Log10(2.0)*256))
			cWaveTable[0x0000+i] = int16((uint16(0x8000) | uint16(cWaveTable[0x200+i])))
		}
		//Exponential wave
		for i := 0; i < 256; i++ {
			cWaveTable[0x700+i] = int16(i * 8)
			cWaveTable[0x6ff-i] = int16(0x8000 | i*8)
		}
	}

	//	|    |//\\|____|WAV7|//__|/\  |____|/\/\|
	//	|\\//|    |    |WAV7|    |  \/|    |    |
	//	|06  |0126|27  |7   |3   |4   |4 5 |5   |

	if cDBOPLWave == cWaveTableLog || cDBOPLWave == cWaveTableMul {
		for i := 0; i < 256; i++ {
			//Fill silence gaps
			cWaveTable[0x400+i] = cWaveTable[0]
			cWaveTable[0x500+i] = cWaveTable[0]
			cWaveTable[0x900+i] = cWaveTable[0]
			cWaveTable[0xc00+i] = cWaveTable[0]
			cWaveTable[0xd00+i] = cWaveTable[0]
			//Replicate sines in other pieces
			cWaveTable[0x800+i] = cWaveTable[0x200+i]
			//float64 speed sines
			cWaveTable[0xa00+i] = cWaveTable[0x200+i*2]
			cWaveTable[0xb00+i] = cWaveTable[0x000+i*2]
			cWaveTable[0xe00+i] = cWaveTable[0x200+i*2]
			cWaveTable[0xf00+i] = cWaveTable[0x200+i*2]
		}
	}

	//Create the ksl table
	for oct := int(0); oct < 8; oct++ {
		base := int(oct * 8)
		for i := 0; i < 16; i++ {
			val := base - int(cKslCreateTable[i])
			if val < 0 {
				val = 0
			}
			//*4 for the final range to match attenuation range
			cKslTable[oct*16+i] = uint8(val * 4)
		}
	}
	//Create the Tremolo table, just increase and decrease a triangle wave
	for i := uint8(0); i < cTremoloTableSize/2; i++ {
		val := uint8(i << cEnvExtra)
		cTremoloTable[i] = val
		cTremoloTable[cTremoloTableSize-1-i] = val
	}
}
