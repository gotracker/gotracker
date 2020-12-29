package load

import (
	"errors"
	"math"

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

func xmInstrumentToInstrument(inst *xmfile.InstrumentHeader) (*layout.Instrument, error) {
	var si *xmfile.SampleHeader
	if inst.SamplesCount > 0 {
		si = &inst.Samples[0]
	} else {
		si = &xmfile.SampleHeader{}
	}
	sample := layout.Instrument{
		Filename:           inst.GetName(),
		Name:               si.GetName(),
		C2Spd:              note.C2SPD(0), // TODO: use si.Finetune
		Volume:             util.VolumeFromXm(si.Volume),
		RelativeNoteNumber: si.RelativeNoteNumber,
	}
	if si.Finetune != 0 {
		n := float64(4 * 12)
		period := 10*12*16*4 - n*16*4 - float64(si.Finetune)/2
		frequency := 8363 * math.Pow(2, ((6*12*16*4-period)/(12*16*4)))
		sample.C2Spd = note.C2SPD(frequency)
	}
	if sample.C2Spd == 0 {
		sample.C2Spd = note.C2SPD(util.DefaultC2Spd)
	}

	idata := layout.InstrumentPCM{
		Length:        int(si.Length),
		Looped:        si.Flags.LoopMode() != xmfile.SampleLoopModeDisabled,
		LoopBegin:     int(si.LoopStart),
		LoopEnd:       int(si.LoopStart + si.LoopLength),
		NumChannels:   1,
		BitsPerSample: 8,
	}
	if si.Flags.IsStereo() {
		idata.NumChannels = 2
	}
	if si.Flags.Is16Bit() {
		idata.BitsPerSample = 16
	}
	stride := idata.NumChannels * idata.BitsPerSample / 8
	idata.Length /= stride
	idata.LoopBegin /= stride
	idata.LoopEnd /= stride

	idata.Sample = make([]uint8, len(si.SampleData))
	for i, s := range si.SampleData {
		idata.Sample[i] = uint8(s)
	}

	sample.Inst = &idata
	return &sample, nil
}

func convertXMInstrumentToInstrument(s *xmfile.InstrumentHeader) (*layout.Instrument, error) {
	if s == nil {
		return nil, errors.New("instrument is nil")
	}

	return xmInstrumentToInstrument(s)
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

	song := layout.Song{
		Head:        *h,
		Instruments: make([]layout.Instrument, len(f.Instruments)),
		Patterns:    make([]layout.Pattern, len(f.Patterns)),
		OrderList:   make([]intf.PatternIdx, int(f.Head.SongLength)),
	}

	for i := 0; i < int(f.Head.SongLength); i++ {
		song.OrderList[i] = intf.PatternIdx(f.Head.OrderTable[i])
	}

	song.Instruments = make([]layout.Instrument, len(f.Instruments))
	for instNum, scrs := range f.Instruments {
		sample, err := convertXMInstrumentToInstrument(&scrs)
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

	channels := make([]layout.ChannelSetting, lastEnabledChannel)
	for chNum := range channels {
		cs := layout.ChannelSetting{
			Enabled:        true,
			InitialVolume:  util.DefaultVolume,
			InitialPanning: util.DefaultPanning,
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
