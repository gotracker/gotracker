package s3m

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"gotracker/internal/format/s3m/channel"
	"gotracker/internal/format/s3m/util"
	"gotracker/internal/player/note"
	"math"
)

type modHeader struct {
	Name       [20]byte
	Samples    [31]modSample
	SongLen    uint8
	RestartPos uint8
	Order      [128]uint8
	Sig        [4]uint8
}

type modSample struct {
	Name      [22]byte
	Len       uint16
	FineTune  uint8
	Volume    uint8
	LoopStart uint16
	LoopEnd   uint16
}

func readMODHeader(channels int, buffer *bytes.Buffer) (*Header, *modHeader, error) {
	head := Header{}

	mh := modHeader{}
	binary.Read(buffer, binary.LittleEndian, &mh)

	numPatterns := 0
	head.OrderList = make([]uint8, mh.SongLen)
	for i, o := range mh.Order {
		if i < int(mh.SongLen) {
			head.OrderList[i] = o
		}
		if numPatterns-1 < int(o) {
			numPatterns = int(o) + 1
		}
	}

	head.Name = getString(mh.Name[:])
	head.Info.OrderCount = uint16(mh.SongLen)
	head.Info.InstrumentCount = 31
	head.Info.PatternCount = uint16(numPatterns)
	head.Info.InitialSpeed = 6
	head.Info.InitialTempo = 125
	head.Info.MixingVolume = 64
	head.Info.GlobalVolume = 64

	for i := 0; i < 32; i++ {
		if i >= channels {
			head.ChannelSettings[i] = 255
		} else if i%1 == 0 {
			head.ChannelSettings[i] = ChannelSetting(uint8(ChannelIDL1) + uint8(i)>>1)
		} else {
			head.ChannelSettings[i] = ChannelSetting(uint8(ChannelIDR1) + uint8(i)>>1)
		}
	}

	return &head, &mh, nil
}

func readMODPattern(buffer *bytes.Buffer, channels int) *Pattern {
	pattern := Pattern{}

	for r := 0; r < 64; r++ {
		for c := 0; c < channels; c++ {
			var data [4]uint8
			buffer.Read(data[:])
			sampleNumber := (data[0] & 0xF0) | (data[2] >> 4)
			samplePeriod := (uint16(data[0]&0x0F) << 8) | uint16(data[1])
			effect := (data[2] & 0x0F)
			effectParameter := data[3]

			cd := &pattern.Rows[r].Channels[c]
			cd.What = channel.What(c & 0x1F)

			cd.Instrument = sampleNumber

			if samplePeriod != 0 {
				cd.What = cd.What | channel.WhatNote
				cd.Note = modPeriodToNote(samplePeriod * 4)
			}
			cd.Volume = 255
			if effect != 0 || cd.Info != 0 {
				cd.Info = effectParameter
				switch effect {
				case 0xF: // Set Speed / Tempo
					cd.What = cd.What | channel.WhatCommand
					if cd.Info < 0x20 {
						cd.Command = 'A' - '@' // Set Speed
					} else {
						cd.Command = 'T' - '@' // Tempo
					}
				case 0xB: // Pattern Jump
					cd.What = cd.What | channel.WhatCommand
					cd.Command = 'B' - '@'
				case 0xD: // Pattern Break
					cd.What = cd.What | channel.WhatCommand
					cd.Command = 'C' - '@'
				case 0xA: // Volume Slide
					cd.What = cd.What | channel.WhatCommand
					cd.Command = 'D' - '@'
				case 0x2: // Porta Down
					cd.What = cd.What | channel.WhatCommand
					cd.Command = 'E' - '@'
				case 0x1: // Porta Up
					cd.What = cd.What | channel.WhatCommand
					cd.Command = 'F' - '@'
				case 0x3: // Porta to Note
					cd.What = cd.What | channel.WhatCommand
					cd.Command = 'G' - '@'
				case 0x4: // Vibrato
					cd.What = cd.What | channel.WhatCommand
					cd.Command = 'H' - '@'
				case 0x0: // Arpeggio
					cd.What = cd.What | channel.WhatCommand
					cd.Command = 'J' - '@'
				case 0x6: // Vibrato+VolSlide
					cd.What = cd.What | channel.WhatCommand
					cd.Command = 'K' - '@'
				case 0x5: // Porta+VolSlide
					cd.What = cd.What | channel.WhatCommand
					cd.Command = 'L' - '@'
				case 0x9: // Sample Offset
					cd.What = cd.What | channel.WhatCommand
					cd.Command = 'O' - '@'
				case 0x7: // Tremolo
					cd.What = cd.What | channel.WhatCommand
					cd.Command = 'R' - '@'
				case 0xC: // Set Volume
					cd.What = cd.What | channel.WhatVolume
					cd.Volume = cd.Info
				}

				if effect == 0xE {
					// special
					switch effectParameter >> 4 {
					case 0xA: // Fine VolSlide down
						cd.What = cd.What | channel.WhatCommand
						cd.Command = 'D' - '@'
						cd.Info = 0xF0 | (effectParameter & 0x0F)
					case 0xB: // Fine VolSlide up
						cd.What = cd.What | channel.WhatCommand
						cd.Command = 'S' - '@'
						cd.Info = ((effectParameter & 0x0F) << 4) | 0x0F
					case 0x2: // Fine Porta Down
						cd.What = cd.What | channel.WhatCommand
						cd.Command = 'E' - '@'
						cd.Info = 0xF0 | (effectParameter & 0x0F)
					case 0x1: // Fine Porta Up
						cd.What = cd.What | channel.WhatCommand
						cd.Command = 'F' - '@'
						cd.Info = 0xF0 | (effectParameter & 0x0F)
					case 0x9: // Retrig+VolSlide
						cd.What = cd.What | channel.WhatCommand
						cd.Command = 'Q' - '@'
						cd.Info = (effectParameter & 0x0F)
					case 0x0: // Set Filter on/off
						cd.What = cd.What | channel.WhatCommand
						cd.Command = 'S' - '@'
						cd.Info = 0x00 | (effectParameter & 0x0F)
					case 0x3: // Set Glissando on/off
						cd.What = cd.What | channel.WhatCommand
						cd.Command = 'S' - '@'
						cd.Info = 0x10 | (effectParameter & 0x0F)
					case 0x5: // Set FineTune
						cd.What = cd.What | channel.WhatCommand
						cd.Command = 'S' - '@'
						cd.Info = 0x20 | (effectParameter & 0x0F)
					case 0x4: // Set Vibrato Waveform
						cd.What = cd.What | channel.WhatCommand
						cd.Command = 'S' - '@'
						cd.Info = 0x30 | (effectParameter & 0x0F)
					case 0x7: // Set Tremolo Waveform
						cd.What = cd.What | channel.WhatCommand
						cd.Command = 'S' - '@'
						cd.Info = 0x40 | (effectParameter & 0x0F)
					case 0x8: // Set Pan Position
						cd.What = cd.What | channel.WhatCommand
						cd.Command = 'S' - '@'
						cd.Info = 0x80 | (effectParameter & 0x0F)
					case 0x6: // Pattern Loop
						cd.What = cd.What | channel.WhatCommand
						cd.Command = 'S' - '@'
						cd.Info = 0xB0 | (effectParameter & 0x0F)
					case 0xC: // Note Cut
						cd.What = cd.What | channel.WhatCommand
						cd.Command = 'S' - '@'
						cd.Info = 0xC0 | (effectParameter & 0x0F)
					case 0xD: // Note Delay
						cd.What = cd.What | channel.WhatCommand
						cd.Command = 'S' - '@'
						cd.Info = 0xD0 | (effectParameter & 0x0F)
					case 0xE: // Pattern Delay
						cd.What = cd.What | channel.WhatCommand
						cd.Command = 'S' - '@'
						cd.Info = 0xE0 | (effectParameter & 0x0F)
					case 0xF: // Funk Repeat
						cd.What = cd.What | channel.WhatCommand
						cd.Command = 'S' - '@'
						cd.Info = 0xF0 | (effectParameter & 0x0F)
					}
				}
			}
		}
	}

	return &pattern
}

var (
	finetuneC2Spds = [...]note.C2SPD{
		8363, 8413, 8463, 8529, 8581, 8651, 8723, 8757,
		7895, 7941, 7985, 8046, 8107, 8169, 8232, 8280,
	}
)

func readMODSample(buffer *bytes.Buffer, num int, inst modSample) *SampleFileFormat {
	var sample = SampleFileFormat{}
	sample.Filename = fmt.Sprintf("inst%0.2d.bin", num+1)
	sample.Name = getString(inst.Name[:])
	sl := util.BE16ToLE16(inst.Len) * 2

	sample.C2Spd = finetuneC2Spds[inst.FineTune&0xF]

	sample.Volume = util.VolumeFromS3M(inst.Volume)
	sample.LoopBegin = int(util.BE16ToLE16(inst.LoopStart)) * 2
	loopLen := int(util.BE16ToLE16(inst.LoopEnd)) * 2
	sample.LoopEnd = sample.LoopBegin + loopLen
	if loopLen > 2 {
		sample.Looped = true
	}

	samps := make([]uint8, sl)
	buffer.Read(samps)
	sample.Sample = samps
	for i, s := range samps {
		sample.Sample[i] = modSampleToS3MSample(s)
	}
	return &sample
}

type modSig struct {
	sig      string
	channels int
}

var (
	sigChannels = [...]modSig{
		// amiga / protracker
		{"M.K.", 4},
		// fasttracker
		{"6CHN", 6}, {"8CHN", 8},
		// (unusual)
		{"10CH", 10}, {"11CH", 11}, {"12CH", 12}, {"13CH", 13}, {"14CH", 14},
		{"15CH", 15}, {"16CH", 16}, {"17CH", 17}, {"18CH", 18}, {"19CH", 19},
		{"20CH", 20}, {"21CH", 21}, {"22CH", 22}, {"23CH", 23}, {"24CH", 24},
		{"25CH", 25}, {"26CH", 26}, {"27CH", 27}, {"28CH", 28}, {"29CH", 29},
		{"30CH", 30}, {"31CH", 31}, {"32CH", 32},
	}
)

func readMOD(filename string) (*Song, error) {
	buffer, err := readFile(filename)
	if err != nil {
		return nil, err
	}
	data := buffer.Bytes()

	song := Song{}

	sig := getString(data[1080:1084])
	numCh := 0
	for _, s := range sigChannels {
		if s.sig == sig {
			numCh = s.channels
			break
		}
	}

	if numCh == 0 {
		return nil, errors.New("invalid file format")
	}

	h, mh, err := readMODHeader(numCh, buffer)
	if err != nil {
		return nil, err
	}
	song.Head = *h

	song.Patterns = make([]Pattern, song.Head.Info.PatternCount)
	for i := 0; i < int(song.Head.Info.PatternCount); i++ {
		var pattern = readMODPattern(buffer, numCh)
		if pattern == nil {
			continue
		}
		song.Patterns[i] = *pattern
	}

	song.Instruments = make([]SampleFileFormat, len(mh.Samples))
	for instNum, inst := range mh.Samples {
		var sample = readMODSample(buffer, instNum, inst)
		if sample == nil {
			continue
		}
		sample.ID = uint8(instNum + 1)
		song.Instruments[instNum] = *sample
	}

	return &song, nil
}

var (
	modPeriodTable = [...]uint16{
		27392, 25856, 24384, 23040, 21696, 20480, 19328, 18240, 17216, 16256, 15360, 14496,
		13696, 12928, 12192, 11520, 10848, 10240, 9664, 9120, 8608, 8128, 7680, 7248,
		6848, 6464, 6096, 5760, 5424, 5120, 4832, 4560, 4304, 4064, 3840, 3624,
		3424, 3232, 3048, 2880, 2712, 2560, 2416, 2280, 2152, 2032, 1920, 1812,
		1712, 1616, 1524, 1440, 1356, 1280, 1208, 1140, 1076, 1016, 960, 906,
		856, 808, 762, 720, 678, 640, 604, 570, 538, 508, 480, 453,
		428, 404, 381, 360, 339, 320, 302, 285, 269, 254, 240, 226,
		214, 202, 190, 180, 170, 160, 151, 143, 135, 127, 120, 113,
		107, 101, 95, 90, 85, 80, 75, 71, 67, 63, 60, 56,
		// unsupported
		53, 50, 47, 45, 42, 40, 37, 35, 33, 31, 30, 28,
		26, 25, 23, 22, 21, 20, 18, 17, 16, 15, 15, 14,
	}
)

func modPeriodToNote(period uint16) note.Note {
	periodFloat := float64(period)
	for i, pv := range modPeriodTable {
		k := uint8(i % 12)
		o := uint8(i / 12)
		v := math.Abs((periodFloat - float64(pv)) / periodFloat)
		if v < 0.05 {
			return note.Note((o << 4) | (k & 0x0F))
		}
	}
	return note.EmptyNote
}

func modSampleToS3MSample(sample uint8) uint8 {
	return sample - 0x80
}
