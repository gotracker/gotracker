package modfile

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"

	"gotracker/internal/format/s3m/channel"
	"gotracker/internal/format/s3m/s3mfile"
	"gotracker/internal/format/s3m/util"
	"gotracker/internal/player/note"
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

func readMODPattern(buffer *bytes.Buffer, channels int) *s3mfile.PackedPattern {
	w := &bytes.Buffer{}

	for r := 0; r < 64; r++ {
		worthwhileChannels := 0
		unpackedChannels := make([]channel.Data, channels)
		for c := 0; c < channels; c++ {
			var data [4]uint8
			buffer.Read(data[:])
			sampleNumber := (data[0] & 0xF0) | (data[2] >> 4)
			samplePeriod := (uint16(data[0]&0x0F) << 8) | uint16(data[1])
			effect := (data[2] & 0x0F)
			effectParameter := data[3]

			u := &unpackedChannels[c]
			*u = channel.Data{
				What:       channel.What(c & 0x1F),
				Note:       note.EmptyNote,
				Instrument: sampleNumber,
				Volume:     uint8(255),
				Command:    uint8(0),
				Info:       uint8(0),
			}

			if samplePeriod != 0 {
				u.What |= channel.WhatNote
				u.Note = modPeriodToNote(samplePeriod * 4)
			}
			if effect != 0 || effectParameter != 0 {
				u.Info = effectParameter
				switch effect {
				case 0xF: // Set Speed / Tempo
					u.What |= channel.WhatCommand
					if u.Info < 0x20 {
						u.Command = 'A' - '@' // Set Speed
					} else {
						u.Command = 'T' - '@' // Tempo
					}
				case 0xB: // Pattern Jump
					u.What |= channel.WhatCommand
					u.Command = 'B' - '@'
				case 0xD: // Pattern Break
					u.What |= channel.WhatCommand
					u.Command = 'C' - '@'
				case 0xA: // Volume Slide
					u.What |= channel.WhatCommand
					u.Command = 'D' - '@'
				case 0x2: // Porta Down
					u.What |= channel.WhatCommand
					u.Command = 'E' - '@'
				case 0x1: // Porta Up
					u.What |= channel.WhatCommand
					u.Command = 'F' - '@'
				case 0x3: // Porta to Note
					u.What |= channel.WhatCommand
					u.Command = 'G' - '@'
				case 0x4: // Vibrato
					u.What |= channel.WhatCommand
					u.Command = 'H' - '@'
				case 0x0: // Arpeggio
					u.What |= channel.WhatCommand
					u.Command = 'J' - '@'
				case 0x6: // Vibrato+VolSlide
					u.What |= channel.WhatCommand
					u.Command = 'K' - '@'
				case 0x5: // Porta+VolSlide
					u.What |= channel.WhatCommand
					u.Command = 'L' - '@'
				case 0x9: // Sample Offset
					u.What |= channel.WhatCommand
					u.Command = 'O' - '@'
				case 0x7: // Tremolo
					u.What |= channel.WhatCommand
					u.Command = 'R' - '@'
				case 0xC: // Set Volume
					u.What |= channel.WhatVolume
					u.Volume = u.Info
				}

				if effect == 0xE {
					// special
					switch effectParameter >> 4 {
					case 0xA: // Fine VolSlide down
						u.What |= channel.WhatCommand
						u.Command = 'D' - '@'
						u.Info = 0xF0 | (effectParameter & 0x0F)
					case 0xB: // Fine VolSlide up
						u.What |= channel.WhatCommand
						u.Command = 'S' - '@'
						u.Info = ((effectParameter & 0x0F) << 4) | 0x0F
					case 0x2: // Fine Porta Down
						u.What |= channel.WhatCommand
						u.Command = 'E' - '@'
						u.Info = 0xF0 | (effectParameter & 0x0F)
					case 0x1: // Fine Porta Up
						u.What |= channel.WhatCommand
						u.Command = 'F' - '@'
						u.Info = 0xF0 | (effectParameter & 0x0F)
					case 0x9: // Retrig+VolSlide
						u.What |= channel.WhatCommand
						u.Command = 'Q' - '@'
						u.Info = (effectParameter & 0x0F)
					case 0x0: // Set Filter on/off
						u.What |= channel.WhatCommand
						u.Command = 'S' - '@'
						u.Info = 0x00 | (effectParameter & 0x0F)
					case 0x3: // Set Glissando on/off
						u.What |= channel.WhatCommand
						u.Command = 'S' - '@'
						u.Info = 0x10 | (effectParameter & 0x0F)
					case 0x5: // Set FineTune
						u.What |= channel.WhatCommand
						u.Command = 'S' - '@'
						u.Info = 0x20 | (effectParameter & 0x0F)
					case 0x4: // Set Vibrato Waveform
						u.What |= channel.WhatCommand
						u.Command = 'S' - '@'
						u.Info = 0x30 | (effectParameter & 0x0F)
					case 0x7: // Set Tremolo Waveform
						u.What |= channel.WhatCommand
						u.Command = 'S' - '@'
						u.Info = 0x40 | (effectParameter & 0x0F)
					case 0x8: // Set Pan Position
						u.What |= channel.WhatCommand
						u.Command = 'S' - '@'
						u.Info = 0x80 | (effectParameter & 0x0F)
					case 0x6: // Pattern Loop
						u.What |= channel.WhatCommand
						u.Command = 'S' - '@'
						u.Info = 0xB0 | (effectParameter & 0x0F)
					case 0xC: // Note Cut
						u.What |= channel.WhatCommand
						u.Command = 'S' - '@'
						u.Info = 0xC0 | (effectParameter & 0x0F)
					case 0xD: // Note Delay
						u.What |= channel.WhatCommand
						u.Command = 'S' - '@'
						u.Info = 0xD0 | (effectParameter & 0x0F)
					case 0xE: // Pattern Delay
						u.What |= channel.WhatCommand
						u.Command = 'S' - '@'
						u.Info = 0xE0 | (effectParameter & 0x0F)
					case 0xF: // Funk Repeat
						u.What |= channel.WhatCommand
						u.Command = 'S' - '@'
						u.Info = 0xF0 | (effectParameter & 0x0F)
					}
				}
			}

			if u.What.HasNote() || u.What.HasCommand() || u.What.HasVolume() {
				worthwhileChannels = c + 1
			}
		}

		for c, u := range unpackedChannels {
			if c >= worthwhileChannels {
				break
			}

			if u.What == 0 {
				u.What |= channel.WhatNote
				u.Note = note.EmptyNote
				u.Instrument = 0
			}
			binary.Write(w, binary.LittleEndian, u.What)
			if u.What.HasNote() {
				binary.Write(w, binary.LittleEndian, u.Note)
				binary.Write(w, binary.LittleEndian, u.Instrument)
			}
			if u.What.HasVolume() {
				binary.Write(w, binary.LittleEndian, u.Volume)
			}
			if u.What.HasCommand() {
				binary.Write(w, binary.LittleEndian, u.Command)
				binary.Write(w, binary.LittleEndian, u.Info)
			}
		}
		binary.Write(w, binary.LittleEndian, uint8(0))
	}

	pattern := s3mfile.PackedPattern{
		Length: uint16(w.Len() + 2),
		Data:   w.Bytes(),
	}

	return &pattern
}

var (
	finetuneC2Spds = [...]note.C2SPD{
		8363, 8413, 8463, 8529, 8581, 8651, 8723, 8757,
		7895, 7941, 7985, 8046, 8107, 8169, 8232, 8280,
	}
)

func readMODSample(buffer *bytes.Buffer, num int, inst modSample) (*s3mfile.SCRSFull, error) {
	sl := util.BE16ToLE16(inst.Len) * 2
	loopLen := util.BE16ToLE16(inst.LoopEnd) * 2
	anc := s3mfile.SCRSDigiplayerHeader{
		Length: s3mfile.HiLo32{
			Lo: sl,
		},
		C2Spd: s3mfile.HiLo32{
			Lo: uint16(finetuneC2Spds[inst.FineTune&0xF]),
		},
		Volume: inst.Volume,
		LoopBegin: s3mfile.HiLo32{
			Lo: util.BE16ToLE16(inst.LoopStart) * 2,
		},
	}
	anc.LoopEnd.Lo = anc.LoopBegin.Lo + loopLen
	if loopLen > 2 {
		anc.Flags |= s3mfile.SCRSFlagsLooped
	}
	scrs := s3mfile.SCRSFull{
		SCRS: s3mfile.SCRS{
			Head: s3mfile.SCRSHeader{
				Type:     s3mfile.SCRSTypeDigiplayer,
				Filename: [12]byte{'i', 'n', 's', 't', '0' + byte(num+1)/10, '0' + byte(num+1)%10, '.', 'b', 'i', 'n'},
			},
			Ancillary: &anc,
		},
	}
	copy(anc.SampleName[:], inst.Name[:])

	samps := make([]uint8, sl)
	buffer.Read(samps)
	scrs.Sample = samps
	for i, s := range samps {
		samps[i] = modSampleToS3MSample(s)
	}
	return &scrs, nil
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

// Read reads a MOD file from the reader `r` and creates an internal S3M File representation
func Read(r io.Reader) (*s3mfile.File, error) {
	buffer := &bytes.Buffer{}
	if _, err := buffer.ReadFrom(r); err != nil {
		return nil, err
	}
	data := buffer.Bytes()

	sig := util.GetString(data[1080:1084])
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

	f := s3mfile.File{}

	mh := modHeader{}
	if err := binary.Read(buffer, binary.LittleEndian, &mh); err != nil {
		return nil, err
	}

	numPatterns := 0
	orderList := make([]uint8, mh.SongLen)
	for i, o := range mh.Order {
		if i < int(mh.SongLen) {
			orderList[i] = o
		}
		if numPatterns-1 < int(o) {
			numPatterns = int(o) + 1
		}
	}

	f.Head = s3mfile.ModuleHeader{
		Name:            [28]byte{},
		OrderCount:      uint16(mh.SongLen),
		InstrumentCount: 31,
		PatternCount:    uint16(numPatterns),
		GlobalVolume:    64,
		InitialSpeed:    6,
		InitialTempo:    125,
		MixingVolume:    uint8(64 / numCh),
	}

	copy(f.Head.Name[:], mh.Name[:])

	f.OrderList = orderList

	for i := 0; i < 32; i++ {
		if i >= numCh {
			f.ChannelSettings[i] = 255
			continue
		}

		isLeft := (i & 1) == 0
		if isLeft {
			f.ChannelSettings[i] = s3mfile.MakeChannelSetting(true, s3mfile.ChannelCategoryPCMLeft, i>>1)
			f.Panning[i] = s3mfile.DefaultPanningLeft
		} else {
			f.ChannelSettings[i] = s3mfile.MakeChannelSetting(true, s3mfile.ChannelCategoryPCMRight, i>>1)
			f.Panning[i] = s3mfile.DefaultPanningRight
		}

	}

	f.Patterns = make([]s3mfile.PackedPattern, f.Head.PatternCount)
	for i := 0; i < int(f.Head.PatternCount); i++ {
		var pattern = readMODPattern(buffer, numCh)
		if pattern == nil {
			continue
		}
		f.Patterns[i] = *pattern
	}

	f.Instruments = make([]s3mfile.SCRSFull, len(mh.Samples))
	for instNum, inst := range mh.Samples {
		scrs, err := readMODSample(buffer, instNum, inst)
		if err != nil {
			return nil, err
		}
		if scrs == nil {
			scrs = &s3mfile.SCRSFull{}
		}
		f.Instruments[instNum] = *scrs
	}

	return &f, nil
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
