package s3m

import (
	"bytes"
	"encoding/binary"
	"errors"

	s3mfile "github.com/heucuva/goaudiofile/music/tracked/s3m"

	"gotracker/internal/format/s3m/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

func moduleHeaderToHeader(fh *s3mfile.ModuleHeader) (*Header, error) {
	if fh == nil {
		return nil, errors.New("file header is nil")
	}
	head := Header{
		Name:         fh.GetName(),
		InitialSpeed: int(fh.InitialSpeed),
		InitialTempo: int(fh.InitialTempo),
		GlobalVolume: util.VolumeFromS3M(fh.GlobalVolume),
		MixingVolume: util.VolumeFromS3M(fh.MixingVolume),
	}
	return &head, nil
}

func scrsNoneToInstrument(scrs *s3mfile.SCRSFull, si *s3mfile.SCRSNoneHeader) (*Instrument, error) {
	sample := Instrument{
		Filename: scrs.Head.GetFilename(),
		Name:     si.GetSampleName(),
		C2Spd:    note.C2SPD(si.C2Spd.Lo),
		Volume:   util.VolumeFromS3M(si.Volume),
	}
	return &sample, nil
}

func scrsDp30ToInstrument(scrs *s3mfile.SCRSFull, si *s3mfile.SCRSDigiplayerHeader) (*Instrument, error) {
	sample := Instrument{
		Filename:      scrs.Head.GetFilename(),
		Name:          si.GetSampleName(),
		Length:        int(si.Length.Lo),
		C2Spd:         note.C2SPD(si.C2Spd.Lo),
		Volume:        util.VolumeFromS3M(si.Volume),
		Looped:        si.Flags.IsLooped(),
		LoopBegin:     int(si.LoopBegin.Lo),
		LoopEnd:       int(si.LoopEnd.Lo),
		NumChannels:   1,
		BitsPerSample: 8,
	}
	if sample.C2Spd == 0 {
		sample.C2Spd = note.C2SPD(s3mfile.DefaultC2Spd)
	}
	if si.Flags.IsStereo() {
		sample.NumChannels = 2
	}
	if si.Flags.Is16BitSample() {
		sample.BitsPerSample = 16
	}

	sample.Sample = scrs.Sample
	return &sample, nil
}

func scrsOpl2ToInstrument(scrs *s3mfile.SCRSFull, si *s3mfile.SCRSAdlibHeader) (*Instrument, error) {
	sample := Instrument{
		Filename: scrs.Head.GetFilename(),
		Name:     si.GetSampleName(),
		C2Spd:    note.C2SPD(si.C2Spd.Lo),
		Volume:   util.VolumeFromS3M(si.Volume),
	}
	// TODO: support for OPL2/Adlib
	//return &sample, nil
	_ = sample // ignore our `sample` value for now
	return nil, errors.New("unsupported type")
}

func convertSCRSFullToInstrument(s *s3mfile.SCRSFull) (*Instrument, error) {
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

func convertS3MPackedPattern(pkt s3mfile.PackedPattern) (*Pattern, int) {
	pattern := &Pattern{
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

func convertS3MFileToSong(f *s3mfile.File) (*Song, error) {
	h, err := moduleHeaderToHeader(&f.Head)
	if err != nil {
		return nil, err
	}

	song := Song{
		Head:        *h,
		Instruments: make([]Instrument, len(f.InstrumentPointers)),
		Patterns:    make([]intf.Pattern, len(f.PatternPointers)),
		OrderList:   f.OrderList,
	}

	song.Instruments = make([]Instrument, len(f.Instruments))
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

	channels := []ChannelSetting{}
	for chNum, ch := range f.ChannelSettings {
		cs := ChannelSetting{
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

func readS3M(filename string) (*Song, error) {
	buffer, err := readFile(filename)
	if err != nil {
		return nil, err
	}

	s, err := s3mfile.Read(buffer)
	if err != nil {
		return nil, err
	}

	return convertS3MFileToSong(s)
}
