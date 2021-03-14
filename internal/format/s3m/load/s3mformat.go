package load

import (
	"bytes"
	"encoding/binary"
	"errors"

	s3mfile "github.com/gotracker/goaudiofile/music/tracked/s3m"
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"
	"github.com/gotracker/voice/fadeout"
	"github.com/gotracker/voice/loop"
	"github.com/gotracker/voice/pcm"

	formatutil "gotracker/internal/format/internal/util"
	"gotracker/internal/format/s3m/layout"
	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/format/settings"
	"gotracker/internal/index"
	"gotracker/internal/instrument"
	"gotracker/internal/player/note"
	"gotracker/internal/player/pattern"
)

func moduleHeaderToHeader(fh *s3mfile.ModuleHeader) (*layout.Header, error) {
	if fh == nil {
		return nil, errors.New("file header is nil")
	}
	head := layout.Header{
		Name:         fh.GetName(),
		InitialSpeed: int(fh.InitialSpeed),
		InitialTempo: int(fh.InitialTempo),
		GlobalVolume: util.VolumeFromS3M(fh.GlobalVolume),
		Stereo:       (fh.MixingVolume & 0x80) != 0,
	}

	z := uint32(fh.MixingVolume & 0x7f)
	if z < 0x10 {
		z = 0x10
	}
	head.MixingVolume = volume.Volume(z) / volume.Volume(0x80)

	return &head, nil
}

func scrsNoneToInstrument(scrs *s3mfile.SCRSFull, si *s3mfile.SCRSNoneHeader) (*instrument.Instrument, error) {
	sample := instrument.Instrument{
		Static: instrument.StaticValues{
			Filename: scrs.Head.GetFilename(),
			Name:     si.GetSampleName(),
			Volume:   util.VolumeFromS3M(si.Volume),
		},
		C2Spd: note.C2SPD(si.C2Spd.Lo),
	}
	return &sample, nil
}

func scrsDp30ToInstrument(scrs *s3mfile.SCRSFull, si *s3mfile.SCRSDigiplayerHeader, signedSamples bool, s *settings.Settings) (*instrument.Instrument, error) {
	sample := instrument.Instrument{
		Static: instrument.StaticValues{
			Filename: scrs.Head.GetFilename(),
			Name:     si.GetSampleName(),
			Volume:   util.VolumeFromS3M(si.Volume),
		},
		C2Spd: note.C2SPD(si.C2Spd.Lo),
	}
	if sample.C2Spd == 0 {
		sample.C2Spd = note.C2SPD(s3mfile.DefaultC2Spd)
	}

	instLen := int(si.Length.Lo)
	numChannels := 1
	format := pcm.SampleDataFormat8BitUnsigned

	sustainMode := loop.ModeDisabled
	sustainSettings := loop.Settings{
		Begin: int(si.LoopBegin.Lo),
		End:   int(si.LoopEnd.Lo),
	}

	idata := instrument.PCM{
		Loop:         &loop.Disabled{},
		Panning:      panning.CenterAhead,
		MixingVolume: volume.Volume(1),
		FadeOut: fadeout.Settings{
			Mode:   fadeout.ModeDisabled,
			Amount: volume.Volume(0),
		},
	}
	if signedSamples {
		format = pcm.SampleDataFormat8BitSigned
	}
	if si.Flags.IsLooped() {
		sustainMode = loop.ModeNormal
	}
	if si.Flags.IsStereo() {
		numChannels = 2
	}
	if si.Flags.Is16BitSample() {
		if signedSamples {
			format = pcm.SampleDataFormat16BitLESigned
		} else {
			format = pcm.SampleDataFormat16BitLEUnsigned
		}
	}

	idata.SustainLoop = loop.NewLoop(sustainMode, sustainSettings)

	sf := format
	if v, ok := s.Get(settings.NamePreferredSampleFormat); ok {
		if val, ok := v.(pcm.SampleDataFormat); ok {
			sf = val
		}
	}
	if sf == format {
		idata.Sample = pcm.NewSample(scrs.Sample, instLen, numChannels, format)
	} else {
		inSample := pcm.NewSample(scrs.Sample, instLen, numChannels, format)
		outSample, err := pcm.ConvertTo(inSample, sf)
		if err != nil {
			return nil, err
		}
		idata.Sample = outSample
	}

	sample.Inst = &idata
	return &sample, nil
}

func scrsOpl2ToInstrument(scrs *s3mfile.SCRSFull, si *s3mfile.SCRSAdlibHeader) (*instrument.Instrument, error) {
	inst := instrument.Instrument{
		Static: instrument.StaticValues{
			Filename: scrs.Head.GetFilename(),
			Name:     si.GetSampleName(),
			Volume:   util.VolumeFromS3M(si.Volume),
		},
		C2Spd: note.C2SPD(si.C2Spd.Lo),
	}

	idata := instrument.OPL2{
		Modulator: instrument.OPL2OperatorData{
			KeyScaleRateSelect:  si.OPL2.ModulatorKeyScaleRateSelect(),
			Sustain:             si.OPL2.ModulatorSustain(),
			Vibrato:             si.OPL2.ModulatorVibrato(),
			Tremolo:             si.OPL2.ModulatorTremolo(),
			FrequencyMultiplier: uint8(si.OPL2.ModulatorFrequencyMultiplier()),
			KeyScaleLevel:       uint8(si.OPL2.ModulatorKeyScaleLevel()),
			Volume:              uint8(si.OPL2.ModulatorVolume()),
			AttackRate:          si.OPL2.ModulatorAttackRate(),
			DecayRate:           si.OPL2.ModulatorDecayRate(),
			SustainLevel:        si.OPL2.ModulatorSustainLevel(),
			ReleaseRate:         si.OPL2.ModulatorReleaseRate(),
			WaveformSelection:   uint8(si.OPL2.ModulatorWaveformSelection()),
		},
		Carrier: instrument.OPL2OperatorData{
			KeyScaleRateSelect:  si.OPL2.CarrierKeyScaleRateSelect(),
			Sustain:             si.OPL2.CarrierSustain(),
			Vibrato:             si.OPL2.CarrierVibrato(),
			Tremolo:             si.OPL2.CarrierTremolo(),
			FrequencyMultiplier: uint8(si.OPL2.CarrierFrequencyMultiplier()),
			KeyScaleLevel:       uint8(si.OPL2.CarrierKeyScaleLevel()),
			Volume:              uint8(si.OPL2.CarrierVolume()),
			AttackRate:          si.OPL2.CarrierAttackRate(),
			DecayRate:           si.OPL2.CarrierDecayRate(),
			SustainLevel:        si.OPL2.CarrierSustainLevel(),
			ReleaseRate:         si.OPL2.CarrierReleaseRate(),
			WaveformSelection:   uint8(si.OPL2.CarrierWaveformSelection()),
		},
		ModulationFeedback: uint8(si.OPL2.ModulationFeedback()),
		AdditiveSynthesis:  si.OPL2.AdditiveSynthesis(),
	}

	inst.Inst = &idata
	return &inst, nil
}

func convertSCRSFullToInstrument(scrs *s3mfile.SCRSFull, signedSamples bool, s *settings.Settings) (*instrument.Instrument, error) {
	if s == nil {
		return nil, errors.New("scrs is nil")
	}

	switch si := scrs.Ancillary.(type) {
	case nil:
		return nil, errors.New("scrs ancillary is nil")
	case *s3mfile.SCRSNoneHeader:
		return scrsNoneToInstrument(scrs, si)
	case *s3mfile.SCRSDigiplayerHeader:
		return scrsDp30ToInstrument(scrs, si, signedSamples, s)
	case *s3mfile.SCRSAdlibHeader:
		return scrsOpl2ToInstrument(scrs, si)
	default:
	}

	return nil, errors.New("unhandled scrs ancillary type")
}

func convertS3MPackedPattern(pkt s3mfile.PackedPattern, numRows uint8) (*pattern.Pattern, int) {
	pat := &pattern.Pattern{
		Orig: pkt,
	}

	buffer := bytes.NewBuffer(pkt.Data)

	rowNum := uint8(0)
	maxCh := uint8(0)
	for rowNum < numRows {
		pat.Rows = append(pat.Rows, pattern.RowData{})
		row := &pat.Rows[rowNum]
		for {
			var what s3mfile.PatternFlags
			if err := binary.Read(buffer, binary.LittleEndian, &what); err != nil {
				panic(err)
			}

			if what == 0 {
				rowNum++
				break
			}

			channelNum := what.Channel()
			for len(row.Channels) <= int(channelNum) {
				row.Channels = append(row.Channels, &channel.Data{})
			}
			temp := row.Channels[channelNum].(*channel.Data)
			if maxCh < channelNum {
				maxCh = channelNum
			}

			temp.What = what
			temp.Note = 0
			temp.Instrument = 0
			temp.Volume = s3mfile.EmptyVolume
			temp.Command = 0
			temp.Info = 0

			if temp.What.HasNote() {
				if err := binary.Read(buffer, binary.LittleEndian, &temp.Note); err != nil {
					panic(err)
				}
				if err := binary.Read(buffer, binary.LittleEndian, &temp.Instrument); err != nil {
					panic(err)
				}
			}

			if temp.What.HasVolume() {
				if err := binary.Read(buffer, binary.LittleEndian, &temp.Volume); err != nil {
					panic(err)
				}
			}

			if temp.What.HasCommand() {
				if err := binary.Read(buffer, binary.LittleEndian, &temp.Command); err != nil {
					panic(err)
				}
				if err := binary.Read(buffer, binary.LittleEndian, &temp.Info); err != nil {
					panic(err)
				}
			}
		}
	}

	return pat, int(maxCh)
}

func convertS3MFileToSong(f *s3mfile.File, getPatternLen func(patNum int) uint8, s *settings.Settings) (*layout.Song, error) {
	h, err := moduleHeaderToHeader(&f.Head)
	if err != nil {
		return nil, err
	}

	song := layout.Song{
		Head:        *h,
		Instruments: make([]instrument.Instrument, len(f.InstrumentPointers)),
		OrderList:   make([]index.Pattern, len(f.OrderList)),
	}

	signedSamples := false
	if f.Head.FileFormatInformation == 1 {
		signedSamples = true
	}

	//st2Vibrato := (f.Head.Flags & 0x0001) != 0
	//st2Tempo := (f.Head.Flags & 0x0002) != 0
	//amigaSlides := (f.Head.Flags & 0x0004) != 0
	//zeroVolOpt := (f.Head.Flags & 0x0008) != 0
	//amigaLimits := (f.Head.Flags & 0x0010) != 0
	sbFilterEnable := (f.Head.Flags & 0x0020) != 0
	st300volSlides := (f.Head.Flags & 0x0040) != 0
	if f.Head.TrackerVersion == 0x1300 {
		st300volSlides = true
	}
	//ptrSpecialIsValid := (f.Head.Flags & 0x0080) != 0

	for i, o := range f.OrderList {
		song.OrderList[i] = index.Pattern(o)
	}

	song.Instruments = make([]instrument.Instrument, len(f.Instruments))
	for instNum, scrs := range f.Instruments {
		sample, err := convertSCRSFullToInstrument(&scrs, signedSamples, s)
		if err != nil {
			return nil, err
		}
		if sample == nil {
			continue
		}
		sample.Static.ID = channel.S3MInstrumentID(uint8(instNum + 1))
		song.Instruments[instNum] = *sample
	}

	lastEnabledChannel := 0
	song.Patterns = make([]pattern.Pattern, len(f.Patterns))
	for patNum, pkt := range f.Patterns {
		pattern, maxCh := convertS3MPackedPattern(pkt, getPatternLen(patNum))
		if pattern == nil {
			continue
		}
		if lastEnabledChannel < maxCh {
			lastEnabledChannel = maxCh
		}
		song.Patterns[patNum] = *pattern
	}

	channels := make([]layout.ChannelSetting, 0)
	for chNum, ch := range f.ChannelSettings {
		chn := ch.GetChannel()
		cs := layout.ChannelSetting{
			Enabled:          ch.IsEnabled(),
			Category:         chn.GetChannelCategory(),
			OutputChannelNum: int(ch.GetChannel() & 0x07),
			InitialVolume:    util.DefaultVolume,
			InitialPanning:   util.DefaultPanning,
			Memory: channel.Memory{
				VolSlideEveryFrame:         st300volSlides,
				LowPassFilterEnable:        sbFilterEnable,
				ResetMemoryAtStartOfOrder0: true,
			},
		}

		cs.Memory.ResetOscillators()

		pf := f.Panning[chNum]
		if pf.IsValid() {
			cs.InitialPanning = util.PanningFromS3M(pf.Value())
		} else {
			switch cs.Category {
			case s3mfile.ChannelCategoryPCMLeft:
				cs.InitialPanning = util.DefaultPanningLeft
				cs.OutputChannelNum = int(chn - s3mfile.ChannelIDL1)
			case s3mfile.ChannelCategoryPCMRight:
				cs.InitialPanning = util.DefaultPanningRight
				cs.OutputChannelNum = int(chn - s3mfile.ChannelIDR1)
			}
		}

		channels = append(channels, cs)
		if cs.Enabled && lastEnabledChannel < chNum {
			lastEnabledChannel = chNum
		}
	}

	song.ChannelSettings = channels[:lastEnabledChannel+1]

	return &song, nil
}

func readS3M(filename string, s *settings.Settings) (*layout.Song, error) {
	buffer, err := formatutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	f, err := s3mfile.Read(buffer)
	if err != nil {
		return nil, err
	}

	return convertS3MFileToSong(f, func(patNum int) uint8 {
		return 64
	}, s)
}
