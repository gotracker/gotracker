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

import "math"

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
	bitsEnvelope = 9

	envelopeMin       = 0
	bitsEnvelopeExtra = bitsEnvelope - 9
	envelopeMax       = 511 << bitsEnvelopeExtra
	envelopeLimit     = (12 * 256) >> (3 - bitsEnvelopeExtra)

	bitsWave      = 10 + (WAVE_PRECISION * 4)
	bitsWaveShift = 32 - bitsWave
	bitsWaveMask  = (1 << bitsWaveShift) - 1

	//Use the same accuracy as the waves
	bitsLFOShift = bitsWaveShift - 10
	//LFO is controlled by our tremolo 256 sample limit
	lfoMax = 256 << bitsLFOShift

	//Attack/decay/release rate counter shift
	bitsRateShift = 24
	bitsRateMask  = (1 << bitsRateShift) - 1

	//Has to fit within 16bit lookuptable
	bitsMulShift = 16
)

func isEnvelopeSilent(x int) bool {
	return x >= envelopeLimit
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
func envelopeSelect(val uint8) (index int, shift int) {
	if val < 13*4 { //Rate 0 - 12
		shift = 12 - int(val>>2)
		index = int(val) & 3
	} else if val < 15*4 { //rate 13 - 14
		shift = 0
		index = int(val) - 12*4
	} else { //rate 15 and up
		shift = 0
		index = 12
	}
	return
}

func getScaledIncreaseEnvelope(val uint8, scale float64) uint32 {
	index, shift := envelopeSelect(val)
	increase := float64(envelopeIncreaseTable[index])
	increaseMul := 1 << (bitsRateShift + bitsEnvelopeExtra - shift - 3)
	unscaledIncrease := increase * float64(increaseMul)
	scaledIncrease := uint32(scale * unscaledIncrease)
	return scaledIncrease
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
			val := float64((0.5 + (math.Pow(2.0, -1.0+float64(255-s)*(1.0/256)))*(1<<bitsMulShift)))
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
		val := uint8(uint16(i) << bitsEnvelopeExtra)
		tremoloTable[i] = val
		tremoloTable[TREMOLO_TABLE-1-i] = val
	}
}
