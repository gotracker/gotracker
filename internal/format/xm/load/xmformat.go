package load

import (
	"errors"

	xmfile "github.com/gotracker/goaudiofile/music/tracked/xm"

	formatutil "gotracker/internal/format/internal/util"
	"gotracker/internal/format/xm/layout"
	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

func moduleHeaderToHeader(fh *xmfile.ModuleHeader) (*layout.Header, error) {
	if fh == nil {
		return nil, errors.New("file header is nil")
	}
	head := layout.Header{
		Name:         fh.GetName(),
		InitialSpeed: int(fh.DefaultSpeed),
		InitialTempo: int(fh.DefaultTempo),
		GlobalVolume: util.DefaultVolume,
		MixingVolume: util.DefaultMixingVolume,
	}
	return &head, nil
}

func xmInstrumentToInstrument(inst *xmfile.InstrumentHeader, linearFrequencySlides bool) ([]*layout.Instrument, map[int][]note.Semitone, error) {
	noteMap := make(map[int][]note.Semitone)

	var instruments []*layout.Instrument

	for _, si := range inst.Samples {
		v := si.Volume & 0x3f
		sample := layout.Instrument{
			Filename:           si.GetName(),
			Name:               inst.GetName(),
			C2Spd:              note.C2SPD(0), // uses si.Finetune, below
			Volume:             util.VolumeFromXm(0x10 + v),
			RelativeNoteNumber: si.RelativeNoteNumber,
		}

		ii := layout.InstrumentPCM{
			Length:        int(si.Length),
			Looped:        si.Flags.LoopMode() != xmfile.SampleLoopModeDisabled,
			LoopBegin:     int(si.LoopStart),
			LoopEnd:       int(si.LoopStart + si.LoopLength),
			NumChannels:   1,
			BitsPerSample: 8,
		}

		if si.Finetune != 0 {
			sample.C2Spd = util.CalcFinetuneC2Spd(util.DefaultC2Spd, note.Finetune(si.Finetune), linearFrequencySlides)
		}
		if sample.C2Spd == 0 {
			sample.C2Spd = note.C2SPD(util.DefaultC2Spd)
		}
		if si.Flags.IsStereo() {
			ii.NumChannels = 2
		}
		if si.Flags.Is16Bit() {
			ii.BitsPerSample = 16
		}
		stride := ii.NumChannels * ii.BitsPerSample / 8
		ii.Length /= stride
		ii.LoopBegin /= stride
		ii.LoopEnd /= stride

		ii.Sample = make([]uint8, len(si.SampleData))
		copy(ii.Sample, si.SampleData)

		sample.Inst = &ii
		instruments = append(instruments, &sample)
	}

	for st, sn := range inst.SampleNumber {
		i := int(sn)
		if i < len(instruments) {
			noteMap[i] = append(noteMap[i], note.Semitone(st))
		}
	}

	return instruments, noteMap, nil
}

func convertXMInstrumentToInstrument(s *xmfile.InstrumentHeader, linearFrequencySlides bool) ([]*layout.Instrument, map[int][]note.Semitone, error) {
	if s == nil {
		return nil, nil, errors.New("instrument is nil")
	}

	return xmInstrumentToInstrument(s, linearFrequencySlides)
}

func convertXmPattern(pkt xmfile.Pattern) (*layout.Pattern, int) {
	pattern := &layout.Pattern{
		Orig: pkt,
	}

	maxCh := uint8(0)
	for rowNum, drow := range pkt.Data {
		pattern.Rows = append(pattern.Rows, layout.RowData{})
		row := &pattern.Rows[rowNum]
		row.Channels = make([]channel.Data, len(drow))
		for channelNum, chn := range drow {
			cd := &row.Channels[channelNum]
			cd.What = chn.Flags
			cd.Note = chn.Note
			cd.Instrument = chn.Instrument
			cd.Volume = chn.Volume
			cd.Effect = chn.Effect
			cd.EffectParameter = chn.EffectParameter
			if maxCh < uint8(channelNum) {
				maxCh = uint8(channelNum)
			}
		}
	}

	return pattern, int(maxCh)
}

func convertXmFileToSong(f *xmfile.File) (*layout.Song, error) {
	h, err := moduleHeaderToHeader(&f.Head)
	if err != nil {
		return nil, err
	}

	linearFrequencySlides := f.Head.Flags.IsLinearSlides()

	song := layout.Song{
		Head:              *h,
		Instruments:       make(map[uint8]*layout.Instrument),
		InstrumentNoteMap: make(map[uint8]map[note.Semitone]*layout.Instrument),
		Patterns:          make([]layout.Pattern, len(f.Patterns)),
		OrderList:         make([]intf.PatternIdx, int(f.Head.SongLength)),
	}

	for i := 0; i < int(f.Head.SongLength); i++ {
		song.OrderList[i] = intf.PatternIdx(f.Head.OrderTable[i])
	}

	for instNum, scrs := range f.Instruments {
		samples, noteMap, err := convertXMInstrumentToInstrument(&scrs, linearFrequencySlides)
		if err != nil {
			return nil, err
		}
		for _, sample := range samples {
			if sample == nil {
				continue
			}
			sample.ID.InstID = uint8(instNum + 1)
			song.Instruments[sample.ID.InstID] = sample
		}
		for i, sts := range noteMap {
			sample := samples[i]
			inm, ok := song.InstrumentNoteMap[sample.ID.InstID]
			if !ok {
				inm = make(map[note.Semitone]*layout.Instrument)
				song.InstrumentNoteMap[sample.ID.InstID] = inm
			}
			for _, st := range sts {
				inm[st] = samples[i]
			}
		}
	}

	lastEnabledChannel := 0
	song.Patterns = make([]layout.Pattern, len(f.Patterns))
	for patNum, pkt := range f.Patterns {
		pattern, maxCh := convertXmPattern(pkt)
		if pattern == nil {
			continue
		}
		if lastEnabledChannel < maxCh {
			lastEnabledChannel = maxCh
		}
		song.Patterns[patNum] = *pattern
	}

	channels := make([]layout.ChannelSetting, lastEnabledChannel+1)
	for chNum := range channels {
		cs := layout.ChannelSetting{
			Enabled:        true,
			InitialVolume:  util.DefaultVolume,
			InitialPanning: util.DefaultPanning,
			Memory: channel.Memory{
				LinearFreqSlides: linearFrequencySlides,
			},
		}

		channels[chNum] = cs
	}

	song.ChannelSettings = channels

	return &song, nil
}

func readXM(filename string) (*layout.Song, error) {
	buffer, err := formatutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	s, err := xmfile.Read(buffer)
	if err != nil {
		return nil, err
	}

	return convertXmFileToSong(s)
}
