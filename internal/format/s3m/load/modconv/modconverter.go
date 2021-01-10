package modconv

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"

	modfile "github.com/gotracker/goaudiofile/music/tracked/mod"
	s3mfile "github.com/gotracker/goaudiofile/music/tracked/s3m"

	"gotracker/internal/format/s3m/layout/channel"
)

func convertMODPatternToS3M(mp *modfile.Pattern) *s3mfile.PackedPattern {
	w := &bytes.Buffer{}

	for _, row := range mp {
		worthwhileChannels := 0
		unpackedChannels := make([]channel.Data, len(row))
		for c, chn := range row {
			sampleNumber := chn.Instrument()
			samplePeriod := chn.Period()
			effect := chn.Effect()
			effectParameter := chn.EffectParameter()

			u := &unpackedChannels[c]
			*u = channel.Data{
				What:       s3mfile.PatternFlags(c & 0x1F),
				Note:       s3mfile.EmptyNote,
				Instrument: channel.S3MInstrumentID(sampleNumber),
				Volume:     s3mfile.EmptyVolume,
				Command:    uint8(0),
				Info:       uint8(0),
			}

			if samplePeriod != 0 {
				u.What |= s3mfile.PatternFlagNote
				u.Note = modPeriodToNote(samplePeriod * 4)
			}
			if effect != 0 || effectParameter != 0 {
				u.Info = effectParameter
				switch effect {
				case 0xF: // Set Speed / Tempo
					u.What |= s3mfile.PatternFlagCommand
					if u.Info < 0x20 {
						u.Command = 'A' - '@' // Set Speed
					} else {
						u.Command = 'T' - '@' // Tempo
					}
				case 0xB: // Pattern Jump
					u.What |= s3mfile.PatternFlagCommand
					u.Command = 'B' - '@'
				case 0xD: // Pattern Break
					u.What |= s3mfile.PatternFlagCommand
					u.Command = 'C' - '@'
				case 0xA: // Volume Slide
					u.What |= s3mfile.PatternFlagCommand
					u.Command = 'D' - '@'
				case 0x2: // Porta Down
					u.What |= s3mfile.PatternFlagCommand
					u.Command = 'E' - '@'
				case 0x1: // Porta Up
					u.What |= s3mfile.PatternFlagCommand
					u.Command = 'F' - '@'
				case 0x3: // Porta to Note
					u.What |= s3mfile.PatternFlagCommand
					u.Command = 'G' - '@'
				case 0x4: // Vibrato
					u.What |= s3mfile.PatternFlagCommand
					u.Command = 'H' - '@'
				case 0x0: // Arpeggio
					u.What |= s3mfile.PatternFlagCommand
					u.Command = 'J' - '@'
				case 0x6: // Vibrato+VolSlide
					u.What |= s3mfile.PatternFlagCommand
					u.Command = 'K' - '@'
				case 0x5: // Porta+VolSlide
					u.What |= s3mfile.PatternFlagCommand
					u.Command = 'L' - '@'
				case 0x9: // Sample Offset
					u.What |= s3mfile.PatternFlagCommand
					u.Command = 'O' - '@'
				case 0x7: // Tremolo
					u.What |= s3mfile.PatternFlagCommand
					u.Command = 'R' - '@'
				case 0xC: // Set Volume
					u.What |= s3mfile.PatternFlagVolume
					u.Volume = s3mfile.Volume(u.Info)
				}

				if effect == 0xE {
					// special
					switch effectParameter >> 4 {
					case 0xA: // Fine VolSlide down
						u.What |= s3mfile.PatternFlagCommand
						u.Command = 'D' - '@'
						u.Info = 0xF0 | (effectParameter & 0x0F)
					case 0xB: // Fine VolSlide up
						u.What |= s3mfile.PatternFlagCommand
						u.Command = 'S' - '@'
						u.Info = ((effectParameter & 0x0F) << 4) | 0x0F
					case 0x2: // Fine Porta Down
						u.What |= s3mfile.PatternFlagCommand
						u.Command = 'E' - '@'
						u.Info = 0xF0 | (effectParameter & 0x0F)
					case 0x1: // Fine Porta Up
						u.What |= s3mfile.PatternFlagCommand
						u.Command = 'F' - '@'
						u.Info = 0xF0 | (effectParameter & 0x0F)
					case 0x9: // Retrig+VolSlide
						u.What |= s3mfile.PatternFlagCommand
						u.Command = 'Q' - '@'
						u.Info = (effectParameter & 0x0F)
					case 0x0: // Set Filter on/off
						u.What |= s3mfile.PatternFlagCommand
						u.Command = 'S' - '@'
						u.Info = 0x00 | (effectParameter & 0x0F)
					case 0x3: // Set Glissando on/off
						u.What |= s3mfile.PatternFlagCommand
						u.Command = 'S' - '@'
						u.Info = 0x10 | (effectParameter & 0x0F)
					case 0x5: // Set FineTune
						u.What |= s3mfile.PatternFlagCommand
						u.Command = 'S' - '@'
						u.Info = 0x20 | (effectParameter & 0x0F)
					case 0x4: // Set Vibrato Waveform
						u.What |= s3mfile.PatternFlagCommand
						u.Command = 'S' - '@'
						u.Info = 0x30 | (effectParameter & 0x0F)
					case 0x7: // Set Tremolo Waveform
						u.What |= s3mfile.PatternFlagCommand
						u.Command = 'S' - '@'
						u.Info = 0x40 | (effectParameter & 0x0F)
					case 0x8: // Set Pan Position
						u.What |= s3mfile.PatternFlagCommand
						u.Command = 'S' - '@'
						u.Info = 0x80 | (effectParameter & 0x0F)
					case 0x6: // Pattern Loop
						u.What |= s3mfile.PatternFlagCommand
						u.Command = 'S' - '@'
						u.Info = 0xB0 | (effectParameter & 0x0F)
					case 0xC: // Note Cut
						u.What |= s3mfile.PatternFlagCommand
						u.Command = 'S' - '@'
						u.Info = 0xC0 | (effectParameter & 0x0F)
					case 0xD: // Note Delay
						u.What |= s3mfile.PatternFlagCommand
						u.Command = 'S' - '@'
						u.Info = 0xD0 | (effectParameter & 0x0F)
					case 0xE: // Pattern Delay
						u.What |= s3mfile.PatternFlagCommand
						u.Command = 'S' - '@'
						u.Info = 0xE0 | (effectParameter & 0x0F)
					case 0xF: // Funk Repeat
						u.What |= s3mfile.PatternFlagCommand
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

			if u.What.HasNote() || u.What.HasCommand() || u.What.HasVolume() {
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
	finetuneC2Spds = [...]s3mfile.C2SPD{
		8363, 8413, 8463, 8529, 8581, 8651, 8723, 8757,
		7895, 7941, 7985, 8046, 8107, 8169, 8232, 8280,
	}
)

var (
	modPeriodTable = [...]modfile.Period{
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

func modPeriodToNote(period modfile.Period) s3mfile.Note {
	periodFloat := float64(period)
	for i, pv := range modPeriodTable {
		k := uint8(i % 12)
		o := uint8(i / 12)
		v := math.Abs((periodFloat - float64(pv)) / periodFloat)
		if v < 0.05 {
			return s3mfile.Note((o << 4) | (k & 0x0F))
		}
	}
	return s3mfile.EmptyNote
}

func convertMODInstrumentToS3M(num int, inst *modfile.InstrumentHeader, samp []uint8) (*s3mfile.SCRSFull, error) {
	loopLen := uint16(inst.LoopEnd.Value())
	anc := s3mfile.SCRSDigiplayerHeader{
		Length: s3mfile.HiLo32{
			Lo: uint16(len(samp)),
		},
		C2Spd: s3mfile.HiLo32{
			Lo: uint16(finetuneC2Spds[inst.FineTune&0xF]),
		},
		Volume: s3mfile.Volume(inst.Volume),
		LoopBegin: s3mfile.HiLo32{
			Lo: uint16(inst.LoopStart.Value()),
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

	scrs.Sample = samp
	for i, s := range samp {
		samp[i] = modSampleToS3MSample(s)
	}
	return &scrs, nil
}

// Read reads a MOD file from the reader `r` and creates an internal S3M File representation
func Read(r io.Reader) (*s3mfile.File, error) {
	mf, err := modfile.Read(r)
	if err != nil {
		return nil, err
	}

	// assume we've got a valid channel number at this pattern:row
	numCh := len(mf.Patterns[0][0])

	f := s3mfile.File{
		Head: s3mfile.ModuleHeader{
			Name:            [28]byte{},
			OrderCount:      uint16(mf.Head.SongLen),
			InstrumentCount: 31,
			PatternCount:    uint16(len(mf.Patterns)),
			GlobalVolume:    s3mfile.DefaultVolume,
			InitialSpeed:    6,
			InitialTempo:    125,
			MixingVolume:    s3mfile.Volume(0x30) | s3mfile.Volume(0x80), // default mixing volume (0x30) for a converted mod in st3, stereo enabled (0x80)
		},
	}

	copy(f.Head.Name[:], mf.Head.Name[:])

	f.OrderList = mf.Head.Order[:int(mf.Head.SongLen)]

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
	for i, p := range mf.Patterns {
		pattern := convertMODPatternToS3M(&p)
		if pattern == nil {
			continue
		}
		f.Patterns[i] = *pattern
	}

	f.Instruments = make([]s3mfile.SCRSFull, len(mf.Samples))
	for instNum, inst := range mf.Head.Instrument {
		scrs, err := convertMODInstrumentToS3M(instNum, &inst, mf.Samples[instNum])
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

func modSampleToS3MSample(sample uint8) uint8 {
	return sample - 0x80
}
