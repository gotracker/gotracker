package load

import (
	"bytes"
	"encoding/binary"
	"errors"

	s3mfile "github.com/heucuva/goaudiofile/music/tracked/s3m"

	formatutil "gotracker/internal/format/internal/util"
	"gotracker/internal/format/s3m/layout"
	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
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
		MixingVolume: util.VolumeFromS3M(fh.MixingVolume),
	}
	return &head, nil
}

func scrsNoneToInstrument(scrs *s3mfile.SCRSFull, si *s3mfile.SCRSNoneHeader) (*layout.Instrument, error) {
	sample := layout.Instrument{
		Filename: scrs.Head.GetFilename(),
		Name:     si.GetSampleName(),
		C2Spd:    note.C2SPD(si.C2Spd.Lo),
		Volume:   util.VolumeFromS3M(si.Volume),
	}
	return &sample, nil
}

func scrsDp30ToInstrument(scrs *s3mfile.SCRSFull, si *s3mfile.SCRSDigiplayerHeader) (*layout.Instrument, error) {
	sample := layout.Instrument{
		Filename: scrs.Head.GetFilename(),
		Name:     si.GetSampleName(),
		C2Spd:    note.C2SPD(si.C2Spd.Lo),
		Volume:   util.VolumeFromS3M(si.Volume),
	}
	if sample.C2Spd == 0 {
		sample.C2Spd = note.C2SPD(s3mfile.DefaultC2Spd)
	}

	idata := layout.InstrumentPCM{
		Length:        int(si.Length.Lo),
		Looped:        si.Flags.IsLooped(),
		LoopBegin:     int(si.LoopBegin.Lo),
		LoopEnd:       int(si.LoopEnd.Lo),
		NumChannels:   1,
		BitsPerSample: 8,
	}
	if si.Flags.IsStereo() {
		idata.NumChannels = 2
	}
	if si.Flags.Is16BitSample() {
		idata.BitsPerSample = 16
	}

	idata.Sample = scrs.Sample

	sample.Inst = &idata
	return &sample, nil
}

func scrsOpl2ToInstrument(scrs *s3mfile.SCRSFull, si *s3mfile.SCRSAdlibHeader) (*layout.Instrument, error) {
	inst := layout.Instrument{
		Filename: scrs.Head.GetFilename(),
		Name:     si.GetSampleName(),
		C2Spd:    note.C2SPD(si.C2Spd.Lo),
		Volume:   util.VolumeFromS3M(si.Volume),
	}

	idata := layout.InstrumentOPL2{
		Modulator: layout.OPL2OperatorData{
			KeyScaleRateSelect:  si.OPL2.ModulatorKeyScaleRateSelect(),
			Sustain:             si.OPL2.ModulatorSustain(),
			Vibrato:             si.OPL2.ModulatorVibrato(),
			Tremolo:             si.OPL2.ModulatorTremolo(),
			FrequencyMultiplier: si.OPL2.ModulatorFrequencyMultiplier(),
			KeyScaleLevel:       si.OPL2.ModulatorKeyScaleLevel(),
			Volume:              s3mfile.Volume(si.OPL2.ModulatorVolume()),
			AttackRate:          si.OPL2.ModulatorAttackRate(),
			DecayRate:           si.OPL2.ModulatorDecayRate(),
			SustainLevel:        si.OPL2.ModulatorSustainLevel(),
			ReleaseRate:         si.OPL2.ModulatorReleaseRate(),
			WaveformSelection:   si.OPL2.ModulatorWaveformSelection(),
		},
		Carrier: layout.OPL2OperatorData{
			KeyScaleRateSelect:  si.OPL2.CarrierKeyScaleRateSelect(),
			Sustain:             si.OPL2.CarrierSustain(),
			Vibrato:             si.OPL2.CarrierVibrato(),
			Tremolo:             si.OPL2.CarrierTremolo(),
			FrequencyMultiplier: si.OPL2.CarrierFrequencyMultiplier(),
			KeyScaleLevel:       si.OPL2.CarrierKeyScaleLevel(),
			Volume:              s3mfile.Volume(si.OPL2.CarrierVolume()),
			AttackRate:          si.OPL2.CarrierAttackRate(),
			DecayRate:           si.OPL2.CarrierDecayRate(),
			SustainLevel:        si.OPL2.CarrierSustainLevel(),
			ReleaseRate:         si.OPL2.CarrierReleaseRate(),
			WaveformSelection:   si.OPL2.CarrierWaveformSelection(),
		},
		ModulationFeedback: si.OPL2.ModulationFeedback(),
		AdditiveSynthesis:  si.OPL2.AdditiveSynthesis(),
	}

	inst.Inst = &idata
	return &inst, nil
}

func convertSCRSFullToInstrument(s *s3mfile.SCRSFull) (*layout.Instrument, error) {
	if s == nil {
		return nil, errors.New("scrs is nil")
	}

	switch si := s.Ancillary.(type) {
	case nil:
		return nil, errors.New("scrs ancillary is nil")
	case *s3mfile.SCRSNoneHeader:
		return scrsNoneToInstrument(s, si)
	case *s3mfile.SCRSDigiplayerHeader:
		return scrsDp30ToInstrument(s, si)
	case *s3mfile.SCRSAdlibHeader:
		return scrsOpl2ToInstrument(s, si)
	default:
	}

	return nil, errors.New("unhandled scrs ancillary type")
}

func convertS3MPackedPattern(pkt s3mfile.PackedPattern) (*layout.Pattern, int) {
	pattern := &layout.Pattern{
		Packed: pkt,
	}

	buffer := bytes.NewBuffer(pkt.Data)

	rowNum := 0
	maxCh := uint8(0)
	for rowNum < len(pattern.Rows) {
		row := &pattern.Rows[rowNum]
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
			temp := &row.Channels[channelNum]
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

	return pattern, int(maxCh)
}

func convertS3MFileToSong(f *s3mfile.File) (*layout.Song, error) {
	h, err := moduleHeaderToHeader(&f.Head)
	if err != nil {
		return nil, err
	}

	song := layout.Song{
		Head:        *h,
		Instruments: make([]layout.Instrument, len(f.InstrumentPointers)),
		Patterns:    make([]intf.Pattern, len(f.PatternPointers)),
		OrderList:   f.OrderList,
	}

	song.Instruments = make([]layout.Instrument, len(f.Instruments))
	for instNum, scrs := range f.Instruments {
		sample, err := convertSCRSFullToInstrument(&scrs)
		if err != nil {
			return nil, err
		}
		if sample == nil {
			continue
		}
		sample.ID = uint8(instNum + 1)
		song.Instruments[instNum] = *sample
	}

	lastEnabledChannel := 0
	song.Patterns = make([]intf.Pattern, len(f.Patterns))
	for patNum, pkt := range f.Patterns {
		pattern, maxCh := convertS3MPackedPattern(pkt)
		if pattern == nil {
			continue
		}
		if lastEnabledChannel < maxCh {
			lastEnabledChannel = maxCh
		}
		song.Patterns[patNum] = pattern
	}

	channels := []layout.ChannelSetting{}
	for chNum, ch := range f.ChannelSettings {
		cs := layout.ChannelSetting{
			Enabled:        ch.IsEnabled(),
			InitialVolume:  util.DefaultVolume,
			InitialPanning: util.DefaultPanning,
		}

		pf := f.Panning[chNum]
		if pf.IsValid() {
			cs.InitialPanning = util.PanningFromS3M(pf.Value())
		} else {
			chn := ch.GetChannel()
			cc := chn.GetChannelCategory()
			switch cc {
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

func readS3M(filename string) (*layout.Song, error) {
	buffer, err := formatutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	s, err := s3mfile.Read(buffer)
	if err != nil {
		return nil, err
	}

	return convertS3MFileToSong(s)
}
