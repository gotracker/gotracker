package load

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"

	itfile "github.com/gotracker/goaudiofile/music/tracked/it"
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/envelope"
	"gotracker/internal/format/it/playback/filter"
	"gotracker/internal/format/it/playback/util"
	"gotracker/internal/instrument"
	"gotracker/internal/loop"
	"gotracker/internal/oscillator"
	"gotracker/internal/pcm"
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
		volEnvLoopMode := loop.ModeDisabled
		volEnvLoopSettings := loop.Settings{
			Begin: int(inst.VolumeLoopStart),
			End:   int(inst.VolumeLoopEnd),
		}
		volEnvSustainMode := loop.ModeDisabled
		volEnvSustainSettings := loop.Settings{
			Begin: int(inst.SustainLoopStart),
			End:   int(inst.SustainLoopEnd),
		}

		id := instrument.PCM{
			Panning: panning.CenterAhead,
			FadeOut: instrument.FadeoutSettings{
				Mode:   instrument.FadeoutModeAlwaysActive,
				Amount: volume.Volume(inst.Fadeout) / 512,
			},
			VolEnv: envelope.Envelope{
				Enabled: (inst.Flags & itfile.IMPIOldFlagUseVolumeEnvelope) != 0,
				Values:  make([]envelope.EnvPoint, 0),
			},
		}

		ii := instrument.Instrument{
			Inst: &id,
		}

		switch inst.NewNoteAction {
		case itfile.NewNoteActionCut:
			ii.NewNoteAction = note.ActionNoteCut
		case itfile.NewNoteActionContinue:
			ii.NewNoteAction = note.ActionContinue
		case itfile.NewNoteActionOff:
			ii.NewNoteAction = note.ActionNoteOff
		case itfile.NewNoteActionFade:
			ii.NewNoteAction = note.ActionFadeout
		default:
			ii.NewNoteAction = note.ActionNoteCut
		}

		ci.Inst = &ii
		addSampleInfoToConvertedInstrument(ci.Inst, &id, &sampData[i], volume.Volume(1), linearFrequencySlides)

		if id.VolEnv.Enabled && id.VolEnv.Loop.Length() >= 0 {
			if enabled := (inst.Flags & itfile.IMPIOldFlagUseVolumeLoop) != 0; enabled {
				volEnvLoopMode = loop.ModeNormal
			}
			if enabled := (inst.Flags & itfile.IMPIOldFlagUseSustainVolumeLoop) != 0; enabled {
				volEnvSustainMode = loop.ModeNormal
			}

			for i := range inst.VolumeEnvelope {
				out := envelope.VolumePoint{}
				in1 := inst.VolumeEnvelope[i]
				vol := volume.Volume(uint8(in1)) / 64
				if vol > 1 {
					vol = 1
				}
				out.Y = vol
				ending := false
				if i+1 >= len(inst.VolumeEnvelope) {
					ending = true
				} else {
					in2 := inst.VolumeEnvelope[i+1]
					if in2 == 0xFF {
						ending = true
					}
				}
				if !ending {
					out.Ticks = 1
				} else {
					out.Ticks = math.MaxInt64
				}
				id.VolEnv.Values = append(id.VolEnv.Values, &out)
			}

			id.VolEnv.Loop = loop.NewLoop(volEnvLoopMode, volEnvLoopSettings)
			id.VolEnv.Sustain = loop.NewLoop(volEnvSustainMode, volEnvSustainSettings)
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
			Panning: panning.CenterAhead,
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
			ii.NewNoteAction = note.ActionNoteCut
		case itfile.NewNoteActionContinue:
			ii.NewNoteAction = note.ActionContinue
		case itfile.NewNoteActionOff:
			ii.NewNoteAction = note.ActionNoteOff
		case itfile.NewNoteActionFade:
			ii.NewNoteAction = note.ActionFadeout
		default:
			ii.NewNoteAction = note.ActionNoteCut
		}

		mixVol := volume.Volume(inst.GlobalVolume.Value())

		ci.Inst = &ii
		addSampleInfoToConvertedInstrument(ci.Inst, &id, &sampData[i], mixVol, linearFrequencySlides)

		var volEnv envelope.VolumePoint
		convertEnvelope(&id.VolEnv, &inst.VolumeEnvelope, reflect.TypeOf(volEnv), func(v int8) interface{} {
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

		var panEnv envelope.PanPoint
		convertEnvelope(&id.PanEnv, &inst.PanningEnvelope, reflect.TypeOf(panEnv), func(v int8) interface{} {
			return panning.MakeStereoPosition(float32(v), -32, 32)
		})

		var pitchEnv envelope.PitchPoint
		id.PitchFiltMode = (inst.PitchEnvelope.Flags & 0x80) != 0 // special flag (IT format changes pitch to resonant filter cutoff envelope)
		convertEnvelope(&id.PitchFiltEnv, &inst.PitchEnvelope, reflect.TypeOf(pitchEnv), func(v int8) interface{} {
			return float32(uint8(v))
		})
	}

	return outInsts, nil
}

func convertEnvelope(outEnv *envelope.Envelope, inEnv *itfile.Envelope, envType reflect.Type, convert func(int8) interface{}) error {
	outEnv.Enabled = (inEnv.Flags & itfile.EnvelopeFlagEnvelopeOn) != 0
	if !outEnv.Enabled {
		return nil
	}

	envLoopMode := loop.ModeDisabled
	envLoopSettings := loop.Settings{
		Begin: int(inEnv.LoopBegin),
		End:   int(inEnv.LoopEnd),
	}
	if enabled := (inEnv.Flags & itfile.EnvelopeFlagLoopOn) != 0; enabled {
		envLoopMode = loop.ModeNormal
	}
	envSustainMode := loop.ModeDisabled
	envSustainSettings := loop.Settings{
		Begin: int(inEnv.SustainLoopBegin),
		End:   int(inEnv.SustainLoopEnd),
	}
	if enabled := (inEnv.Flags & itfile.EnvelopeFlagSustainLoopOn) != 0; enabled {
		envSustainMode = loop.ModeNormal
	}
	outEnv.Values = make([]envelope.EnvPoint, int(inEnv.Count))
	for i := range outEnv.Values {
		in1 := inEnv.NodePoints[i]
		y := convert(in1.Y)
		var ticks int
		if i+1 < len(outEnv.Values) {
			in2 := inEnv.NodePoints[i+1]
			ticks = int(in2.Tick) - int(in1.Tick)
		} else {
			ticks = math.MaxInt64
		}
		out := reflect.New(envType).Interface().(envelope.EnvPoint)
		out.Init(ticks, y)
		outEnv.Values[i] = out
	}

	outEnv.Loop = loop.NewLoop(envLoopMode, envLoopSettings)
	outEnv.Sustain = loop.NewLoop(envSustainMode, envSustainSettings)

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

func getSampleFormat(is16Bit bool, isSigned bool, isBigEndian bool) pcm.SampleDataFormat {
	if is16Bit {
		if isSigned {
			if isBigEndian {
				return pcm.SampleDataFormat16BitBESigned
			}
			return pcm.SampleDataFormat16BitLESigned
		} else if isBigEndian {
			return pcm.SampleDataFormat16BitLEUnsigned
		}
		return pcm.SampleDataFormat16BitLEUnsigned
	} else if isSigned {
		return pcm.SampleDataFormat8BitSigned
	}
	return pcm.SampleDataFormat8BitUnsigned
}

func addSampleInfoToConvertedInstrument(ii *instrument.Instrument, id *instrument.PCM, si *itfile.FullSample, instVol volume.Volume, linearFrequencySlides bool) error {
	instLen := int(si.Header.Length)
	numChannels := 1
	format := pcm.SampleDataFormat8BitUnsigned

	id.MixingVolume = volume.Volume(si.Header.GlobalVolume.Value())
	id.MixingVolume *= instVol
	loopMode := loop.ModeDisabled
	loopSettings := loop.Settings{
		Begin: int(si.Header.LoopBegin),
		End:   int(si.Header.LoopEnd),
	}
	sustainMode := loop.ModeDisabled
	sustainSettings := loop.Settings{
		Begin: int(si.Header.SustainLoopBegin),
		End:   int(si.Header.SustainLoopEnd),
	}

	if si.Header.Flags.IsLoopEnabled() {
		if si.Header.Flags.IsLoopPingPong() {
			loopMode = loop.ModePingPong
		} else {
			loopMode = loop.ModeNormal
		}
	}

	if si.Header.Flags.IsSustainLoopEnabled() {
		if si.Header.Flags.IsSustainLoopPingPong() {
			sustainMode = loop.ModePingPong
		} else {
			sustainMode = loop.ModeNormal
		}
	}

	id.Loop = loop.NewLoop(loopMode, loopSettings)
	id.SustainLoop = loop.NewLoop(sustainMode, sustainSettings)

	if si.Header.Flags.IsStereo() {
		numChannels = 2
	}

	is16Bit := si.Header.Flags.Is16Bit()
	isSigned := si.Header.ConvertFlags.IsSignedSamples()
	isBigEndian := si.Header.ConvertFlags.IsBigEndian()
	format = getSampleFormat(is16Bit, isSigned, isBigEndian)

	isDeltaSamples := si.Header.ConvertFlags.IsSampleDelta()
	var data []byte
	if si.Header.Flags.IsCompressed() {
		if is16Bit {
			data = uncompress16IT214(si.Data, isBigEndian)
		} else {
			data = uncompress8IT214(si.Data)
		}
		isDeltaSamples = true
	} else {
		data = si.Data
	}

	if isDeltaSamples {
		deltaDecode(data, format)
	}

	bytesPerFrame := 2

	if is16Bit {
		bytesPerFrame *= 2
	}

	if len(data) < int(si.Header.Length+1)*bytesPerFrame {
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

		buf := bytes.NewBuffer(data)
		for buf.Len() < int(si.Header.Length+1)*bytesPerFrame {
			binary.Write(buf, order, value)
		}
		data = buf.Bytes()
	}

	id.Sample = pcm.NewSample(data, instLen, numChannels, format)

	ii.Filename = si.Header.GetFilename()
	ii.Name = si.Header.GetName()
	ii.C2Spd = note.C2SPD(si.Header.C5Speed) / note.C2SPD(bytesPerFrame)
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

func deltaDecode(data []byte, format pcm.SampleDataFormat) {
	switch format {
	case pcm.SampleDataFormat8BitSigned, pcm.SampleDataFormat8BitUnsigned:
		deltaDecode8(data)
	case pcm.SampleDataFormat16BitLESigned, pcm.SampleDataFormat16BitLEUnsigned:
		deltaDecode16(data, binary.LittleEndian)
	case pcm.SampleDataFormat16BitBESigned, pcm.SampleDataFormat16BitBEUnsigned:
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
