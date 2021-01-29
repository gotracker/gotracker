package load

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	itfile "github.com/gotracker/goaudiofile/music/tracked/it"
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/format/it/playback/filter"
	"gotracker/internal/format/it/playback/util"
	"gotracker/internal/instrument"
	"gotracker/internal/loop"
	"gotracker/internal/oscillator"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

type convInst struct {
	Inst *instrument.Instrument
	NR   []noteRemap
}

func convertITInstrumentOldToInstrument(inst *itfile.IMPIInstrumentOld, sampData []itfile.FullSample, linearFrequencySlides bool) (map[int]*convInst, error) {
	outInsts := make(map[int]*convInst)

	buildNoteSampleKeyboard(outInsts, inst.NoteSampleKeyboard[:])

	for i, ci := range outInsts {
		id := instrument.PCM{
			NumChannels: 1,
			Format:      instrument.SampleDataFormat8BitUnsigned,
			Panning:     panning.CenterAhead,
			FadeOut: instrument.FadeoutSettings{
				Mode:   instrument.FadeoutModeAlwaysActive,
				Amount: volume.Volume(inst.Fadeout) / 512,
			},
			VolEnv: instrument.InstEnv{
				Enabled:          (inst.Flags & itfile.IMPIOldFlagUseVolumeEnvelope) != 0,
				LoopEnabled:      (inst.Flags & itfile.IMPIOldFlagUseVolumeLoop) != 0,
				SustainEnabled:   (inst.Flags & itfile.IMPIOldFlagUseSustainVolumeLoop) != 0,
				LoopStart:        int(inst.VolumeLoopStart),
				LoopEnd:          int(inst.VolumeLoopEnd),
				SustainLoopStart: int(inst.SustainLoopStart),
				SustainLoopEnd:   int(inst.SustainLoopEnd),
				Values:           make([]instrument.EnvPoint, 0),
			},
		}

		ii := instrument.Instrument{
			Inst: &id,
		}

		switch inst.NewNoteAction {
		case itfile.NewNoteActionCut:
			ii.NewNoteAction = note.NewNoteActionNoteCut
		case itfile.NewNoteActionContinue:
			ii.NewNoteAction = note.NewNoteActionContinue
		case itfile.NewNoteActionOff:
			ii.NewNoteAction = note.NewNoteActionNoteOff
		case itfile.NewNoteActionFade:
			ii.NewNoteAction = note.NewNoteActionFadeout
		default:
			ii.NewNoteAction = note.NewNoteActionNoteCut
		}

		ci.Inst = &ii
		addSampleInfoToConvertedInstrument(ci.Inst, &id, &sampData[i], volume.Volume(1), linearFrequencySlides)

		for i := range inst.VolumeEnvelope {
			out := instrument.EnvPoint{}
			in1 := inst.VolumeEnvelope[i]
			vol := volume.Volume(uint8(in1)) / 64
			if vol > 1 {
				vol = 1
			}
			out.Y0 = vol
			ending := false
			if i+1 >= len(inst.VolumeEnvelope) {
				ending = true
				out.Y1 = volume.Volume(0)
			} else {
				in2 := inst.VolumeEnvelope[i+1]
				if in2 == 0xFF {
					ending = true
					out.Y1 = volume.Volume(0)
				} else {
					out.Y1 = volume.Volume(uint8(in2)) / 64
				}
			}
			if !ending {
				out.Length = 1
			} else {
				out.Length = math.MaxInt64
			}
			id.VolEnv.Values = append(id.VolEnv.Values, out)
		}
	}

	return outInsts, nil
}

func convertITInstrumentToInstrument(inst *itfile.IMPIInstrument, sampData []itfile.FullSample, linearFrequencySlides bool) (map[int]*convInst, error) {
	outInsts := make(map[int]*convInst)

	buildNoteSampleKeyboard(outInsts, inst.NoteSampleKeyboard[:])

	var channelFilterFactory func(sampleRate float32) intf.Filter
	if inst.InitialFilterResonance != 0 {
		channelFilterFactory = func(sampleRate float32) intf.Filter {
			return filter.NewResonantFilter(inst.InitialFilterCutoff, inst.InitialFilterResonance, sampleRate)
		}
	}

	for i, ci := range outInsts {
		id := instrument.PCM{
			NumChannels: 1,
			Format:      instrument.SampleDataFormat8BitUnsigned,
			Panning:     panning.CenterAhead,
			FadeOut: instrument.FadeoutSettings{
				Mode:   instrument.FadeoutModeAlwaysActive,
				Amount: volume.Volume(inst.Fadeout) / 1024,
			},
		}

		ii := instrument.Instrument{
			Inst:                 &id,
			ChannelFilterFactory: channelFilterFactory,
		}

		switch inst.NewNoteAction {
		case itfile.NewNoteActionCut:
			ii.NewNoteAction = note.NewNoteActionNoteCut
		case itfile.NewNoteActionContinue:
			ii.NewNoteAction = note.NewNoteActionContinue
		case itfile.NewNoteActionOff:
			ii.NewNoteAction = note.NewNoteActionNoteOff
		case itfile.NewNoteActionFade:
			ii.NewNoteAction = note.NewNoteActionFadeout
		default:
			ii.NewNoteAction = note.NewNoteActionNoteCut
		}

		mixVol := volume.Volume(inst.GlobalVolume.Value())

		ci.Inst = &ii
		addSampleInfoToConvertedInstrument(ci.Inst, &id, &sampData[i], mixVol, linearFrequencySlides)

		convertEnvelope(&id.VolEnv, &inst.VolumeEnvelope, func(v int8) interface{} {
			vol := volume.Volume(uint8(v)) / 64
			if vol > 1 {
				// NOTE: there might be an incoming Y value == 0xFF, which really
				// means "end of envelope" and should not mean "full volume",
				// but we can cheat a little here and probably get away with it...
				vol = 1
			}
			return vol
		})
		id.VolEnv.OnFinished = func(ioc intf.NoteControl) {
			ioc.Fadeout()
		}

		convertEnvelope(&id.PanEnv, &inst.PanningEnvelope, func(v int8) interface{} {
			return panning.MakeStereoPosition(float32(v), -32, 32)
		})

		convertEnvelope(&id.PitchEnv, &inst.PitchEnvelope, func(v int8) interface{} {
			return note.PeriodDelta(v)
		})
	}

	return outInsts, nil
}

func convertEnvelope(outEnv *instrument.InstEnv, inEnv *itfile.Envelope, convert func(int8) interface{}) error {
	outEnv.Enabled = (inEnv.Flags & itfile.EnvelopeFlagEnvelopeOn) != 0
	outEnv.LoopEnabled = (inEnv.Flags & itfile.EnvelopeFlagLoopOn) != 0
	outEnv.SustainEnabled = (inEnv.Flags & itfile.EnvelopeFlagSustainLoopOn) != 0
	outEnv.LoopStart = int(inEnv.LoopBegin)
	outEnv.LoopEnd = int(inEnv.LoopEnd)
	outEnv.SustainLoopStart = int(inEnv.SustainLoopBegin)
	outEnv.SustainLoopEnd = int(inEnv.SustainLoopEnd)
	outEnv.Values = make([]instrument.EnvPoint, int(inEnv.Count))
	var oldY0 interface{}
	for i := range outEnv.Values {
		out := &outEnv.Values[i]
		in1 := inEnv.NodePoints[i]
		out.Y0 = convert(in1.Y)
		if i+1 < len(outEnv.Values) {
			in2 := inEnv.NodePoints[i+1]
			out.Length = int(in2.Tick) - int(in1.Tick)
			out.Y1 = convert(in2.Y)
		} else {
			out.Length = math.MaxInt64
			if oldY0 != nil {
				out.Y1 = oldY0
			} else {
				out.Y1 = convert(0)
			}
		}
		oldY0 = out.Y0
	}

	return nil
}

func buildNoteSampleKeyboard(noteKeyboard map[int]*convInst, nsk []itfile.NoteSample) error {
	for o, ns := range nsk {
		s := int(ns.Sample)
		if s == 0 {
			continue
		}
		si := int(ns.Sample) - 1
		if si < 0 {
			continue
		}
		n := util.NoteFromItNote(ns.Note)
		st := n.Semitone()
		ci, ok := noteKeyboard[si]
		if !ok {
			ci = &convInst{}
			noteKeyboard[si] = ci
		}
		ci.NR = append(ci.NR, noteRemap{
			Orig:  note.Semitone(o),
			Remap: st,
		})
	}

	return nil
}

func getSampleFormat(is16Bit bool, isSigned bool, isBigEndian bool) instrument.SampleDataFormat {
	if is16Bit {
		if isSigned {
			if isBigEndian {
				return instrument.SampleDataFormat16BitBESigned
			}
			return instrument.SampleDataFormat16BitLESigned
		} else if isBigEndian {
			return instrument.SampleDataFormat16BitLEUnsigned
		}
		return instrument.SampleDataFormat16BitLEUnsigned
	} else if isSigned {
		return instrument.SampleDataFormat8BitSigned
	}
	return instrument.SampleDataFormat8BitUnsigned
}

func addSampleInfoToConvertedInstrument(ii *instrument.Instrument, id *instrument.PCM, si *itfile.FullSample, instVol volume.Volume, linearFrequencySlides bool) error {
	id.Length = int(si.Header.Length)
	id.MixingVolume = volume.Volume(si.Header.GlobalVolume.Value())
	id.MixingVolume *= instVol
	id.Loop = loop.Loop{
		Mode:  loop.ModeDisabled,
		Begin: int(si.Header.LoopBegin),
		End:   int(si.Header.LoopEnd),
	}
	id.SustainLoop = loop.Loop{
		Mode:  loop.ModeDisabled,
		Begin: int(si.Header.SustainLoopBegin),
		End:   int(si.Header.SustainLoopEnd),
	}

	if si.Header.Flags.IsLoopEnabled() {
		if si.Header.Flags.IsLoopPingPong() {
			id.Loop.Mode = loop.ModePingPong
		} else {
			id.Loop.Mode = loop.ModeNormal
		}
	}

	if si.Header.Flags.IsSustainLoopEnabled() {
		if si.Header.Flags.IsSustainLoopPingPong() {
			id.Loop.Mode = loop.ModePingPong
		} else {
			id.Loop.Mode = loop.ModeNormal
		}
	}

	if si.Header.Flags.IsStereo() {
		id.NumChannels = 2
	}

	is16Bit := si.Header.Flags.Is16Bit()
	isSigned := si.Header.ConvertFlags.IsSignedSamples()
	isBigEndian := si.Header.ConvertFlags.IsBigEndian()
	id.Format = getSampleFormat(is16Bit, isSigned, isBigEndian)

	isDeltaSamples := si.Header.ConvertFlags.IsSampleDelta()
	if si.Header.Flags.IsCompressed() {
		if is16Bit {
			id.Sample = uncompress16IT214(si.Data, isBigEndian)
		} else {
			id.Sample = uncompress8IT214(si.Data)
		}
		isDeltaSamples = true
	} else {
		id.Sample = si.Data
	}

	if isDeltaSamples {
		deltaDecode(id.Sample, id.Format)
	}

	bytesPerFrame := 2

	if is16Bit {
		bytesPerFrame *= 2
	}

	if len(id.Sample) < int(si.Header.Length+1)*bytesPerFrame {
		var value interface{}
		var order binary.ByteOrder = binary.LittleEndian
		if is16Bit {
			if isSigned {
				value = int16(0)
			} else {
				value = uint16(0x8000)
			}
			if isBigEndian {
				order = binary.BigEndian
			}
		} else {
			if isSigned {
				value = int8(0)
			} else {
				value = uint8(0x80)
			}
		}

		buf := bytes.NewBuffer(id.Sample)
		for buf.Len() < int(si.Header.Length+1)*bytesPerFrame {
			binary.Write(buf, order, value)
		}
		id.Sample = buf.Bytes()
	}

	ii.Filename = si.Header.GetFilename()
	ii.Name = si.Header.GetName()
	ii.C2Spd = note.C2SPD(si.Header.C5Speed / uint32(bytesPerFrame))
	ii.AutoVibrato = instrument.AutoVibrato{
		Enabled:           (si.Header.VibratoDepth != 0 && si.Header.VibratoSpeed != 0 && si.Header.VibratoSweep != 0),
		Sweep:             0,
		WaveformSelection: si.Header.VibratoType,
		Depth:             si.Header.VibratoDepth,
		Rate:              si.Header.VibratoSpeed,
		Factory: func() oscillator.Oscillator {
			return oscillator.NewImpulseTrackerOscillator(1)
		},
	}
	ii.Volume = volume.Volume(si.Header.Volume.Value())

	if si.Header.VibratoSweep != 0 {
		ii.AutoVibrato.Sweep = uint8(int(si.Header.VibratoDepth) * 256 / int(si.Header.VibratoSweep))
	}
	if !si.Header.DefaultPan.IsDisabled() {
		id.Panning = panning.MakeStereoPosition(si.Header.DefaultPan.Value(), 0, 1)
	}

	return nil
}

func itReadbits(n int8, r io.ByteReader, bitnum *uint32, bitbuf *uint32) (uint32, error) {
	var value uint32 = 0
	var i uint32 = uint32(n)

	// this could be better
	for i > 0 {
		i--
		if *bitnum == 0 {
			b, err := r.ReadByte()
			if err != nil {
				return value >> (32 - n), err
			}
			*bitbuf = uint32(b)
			*bitnum = 8
		}
		value >>= 1
		value |= (*bitbuf) << 31
		(*bitbuf) >>= 1
		(*bitnum)--
	}
	return value >> (32 - n), nil
}

// 8-bit sample uncompressor for IT 2.14+
func uncompress8IT214(data []byte) []byte {
	in := bytes.NewReader(data)
	out := &bytes.Buffer{}

	var (
		blklen uint16 // length of compressed data block in samples
		blkpos uint16 // position in block
		width  uint8  // actual "bit width"
		value  uint16 // value read from file to be processed
		v      int8   // sample value

		// state for itReadbits
		bitbuf uint32
		bitnum uint32
	)

	// now unpack data till the dest buffer is full
	for in.Len() > 0 {
		// read a new block of compressed data and reset variables
		// block layout: word size, <size> bytes data
		bitbuf = 0
		bitnum = 0

		blklen = uint16(math.Min(0x8000, float64(in.Len())))
		blkpos = 0

		width = 9 // start with width of 9 bits

		var clen uint16
		if err := binary.Read(in, binary.LittleEndian, &clen); err != nil {
			panic(err)
		}

		// now uncompress the data block
	blockLoop:
		for blkpos < blklen {
			if width > 9 {
				// illegal width, abort
				panic(fmt.Sprintf("Illegal bit width %d for 8-bit sample\n", width))
			}
			vv, err := itReadbits(int8(width), in, &bitnum, &bitbuf)
			if err != nil {
				break blockLoop
			}
			value = uint16(vv)

			if width < 7 {
				// method 1 (1-6 bits)
				// check for "100..."
				if value == 1<<(width-1) {
					// yes!
					vv, err := itReadbits(3, in, &bitnum, &bitbuf) // read new width
					if err != nil {
						break blockLoop
					}
					value = uint16(vv + 1)
					if value < uint16(width) {
						width = uint8(value)
					} else {
						width = uint8(value + 1)
					}
					continue blockLoop // ... next value
				}
			} else if width < 9 {
				// method 2 (7-8 bits)
				var border uint8 = (0xFF >> (9 - width)) - 4 // lower border for width chg
				if value > uint16(border) && value <= (uint16(border)+8) {
					value -= uint16(border) // convert width to 1-8
					if value < uint16(width) {
						width = uint8(value)
					} else {
						width = uint8(value + 1)
					}
					continue blockLoop // ... next value
				}
			} else {
				// method 3 (9 bits)
				// bit 8 set?
				if (value & 0x100) != 0 {
					width = uint8((value + 1) & 0xff) // new width...
					continue blockLoop                // ... next value
				}
			}

			// now expand value to signed byte
			if width < 8 {
				var shift uint8 = 8 - width
				v = int8(value << shift)
				v >>= shift
			} else {
				v = int8(value)
			}

			if err := out.WriteByte(byte(v)); err != nil {
				panic(err)
			}
			blkpos++
		}
	}
	return out.Bytes()
}

// 16-bit sample uncompressor for IT 2.14+
func uncompress16IT214(data []byte, isBigEndian bool) []byte {
	in := bytes.NewReader(data)
	out := &bytes.Buffer{}

	var (
		blklen uint16 // length of compressed data block in samples
		blkpos uint16 // position in block
		width  uint8  // actual "bit width"
		value  uint32 // value read from file to be processed
		v      int16  // sample value
		order  binary.ByteOrder

		// state for itReadbits
		bitbuf uint32
		bitnum uint32
	)

	if isBigEndian {
		order = binary.BigEndian
	} else {
		order = binary.LittleEndian
	}

	// now unpack data till the dest buffer is full
	for in.Len() > 0 {
		// read a new block of compressed data and reset variables
		// block layout: word size, <size> bytes data
		bitbuf = 0
		bitnum = 0

		blklen = uint16(math.Min(0x4000, float64(in.Len())))
		blkpos = 0

		width = 17 // start with width of 17 bits

		var clen uint16
		if err := binary.Read(in, binary.LittleEndian, &clen); err != nil {
			panic(err)
		}

		// now uncompress the data block
	blockLoop:
		for blkpos < blklen {
			if width > 17 {
				// illegal width, abort
				panic(fmt.Sprintf("Illegal bit width %d for 16-bit sample\n", width))
			}
			vv, err := itReadbits(int8(width), in, &bitnum, &bitbuf)
			if err != nil {
				break blockLoop
			}
			value = vv

			if width < 7 {
				// method 1 (1-6 bits)
				// check for "100..."
				if value == 1<<(width-1) {
					// yes!
					vv, err := itReadbits(4, in, &bitnum, &bitbuf) // read new width
					if err != nil {
						break blockLoop
					}
					value = vv + 1
					if value < uint32(width) {
						width = uint8(value)
					} else {
						width = uint8(value + 1)
					}
					continue blockLoop // ... next value
				}
			} else if width < 17 {
				// method 2 (7-16 bits)
				var border uint16 = (0xFFFF >> (17 - width)) - 8 // lower border for width chg
				if value > uint32(border) && value <= uint32(border+16) {
					value -= uint32(border) // convert width to 1-16
					if value < uint32(width) {
						width = uint8(value)
					} else {
						width = uint8(value + 1)
					}
					continue blockLoop // ... next value
				}
			} else {
				// method 3 (9 bits)
				// bit 8 set?
				if (value & 0x10000) != 0 {
					width = uint8((value + 1) & 0xff) // new width...
					continue blockLoop                // ... next value
				}
			}

			// now expand value to signed byte
			if width < 8 {
				var shift uint8 = 16 - width
				v = int16(value << shift)
				v >>= shift
			} else {
				v = int16(value)
			}

			if err := binary.Write(out, order, v); err != nil {
				panic(err)
			}
			blkpos++
		}
	}
	return out.Bytes()
}

func deltaDecode(data []byte, format instrument.SampleDataFormat) {
	switch format {
	case instrument.SampleDataFormat8BitSigned, instrument.SampleDataFormat8BitUnsigned:
		deltaDecode8(data)
	case instrument.SampleDataFormat16BitLESigned, instrument.SampleDataFormat16BitLEUnsigned:
		deltaDecode16(data, binary.LittleEndian)
	case instrument.SampleDataFormat16BitBESigned, instrument.SampleDataFormat16BitBEUnsigned:
		deltaDecode16(data, binary.BigEndian)
	}
}

func deltaDecode8(data []byte) {
	old := int8(0)
	for i, s := range data {
		new := int8(s) + old
		data[i] = uint8(new)
		old = new
	}
}

func deltaDecode16(data []byte, order binary.ByteOrder) {
	old := int16(0)
	for i := 0; i < len(data); i += 2 {
		s := order.Uint16(data[i:])
		new := int16(s) + old
		order.PutUint16(data[i:], uint16(new))
		old = new
	}
}
