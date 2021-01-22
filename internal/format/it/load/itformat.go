package load

import (
	"bytes"
	"errors"
	"fmt"

	itfile "github.com/gotracker/goaudiofile/music/tracked/it"
	"github.com/gotracker/gomixing/volume"

	formatutil "gotracker/internal/format/internal/util"
	"gotracker/internal/format/it/layout"
	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/format/it/playback/util"
	"gotracker/internal/instrument"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/pattern"
)

func moduleHeaderToHeader(fh *itfile.ModuleHeader) (*layout.Header, error) {
	if fh == nil {
		return nil, errors.New("file header is nil")
	}
	head := layout.Header{
		Name:         fh.GetName(),
		InitialSpeed: int(fh.InitialSpeed),
		InitialTempo: int(fh.InitialTempo),
		GlobalVolume: volume.Volume(fh.GlobalVolume.Value()),
		MixingVolume: volume.Volume(fh.MixingVolume.Value()),
	}
	return &head, nil
}

/*

func itInstrumentToInstrument(inst itfile.IMPIIntf, linearFrequencySlides bool) ([]*instrument.Instrument, map[int][]note.Semitone, error) {
	noteMap := make(map[int][]note.Semitone)

	var instruments []*instrument.Instrument

	for _, si := range inst.Samples {
		v := si.Volume
		if v >= 0x40 {
			v = 0x40
		}
		sample := instrument.Instrument{
			Filename:           si.GetName(),
			Name:               inst.GetName(),
			C2Spd:              note.C2SPD(0), // uses si.Finetune, below
			Volume:             util.VolumeFromIt(0x10 + v),
			RelativeNoteNumber: si.RelativeNoteNumber,
			AutoVibrato: instrument.AutoVibrato{
				Enabled:           (inst.VibratoDepth != 0 && inst.VibratoRate != 0),
				Sweep:             inst.VibratoSweep, // NOTE: for IT support, this needs to be calculated as (Depth * 256 / VibratoSweep) ticks
				WaveformSelection: inst.VibratoType,
				Depth:             inst.VibratoDepth,
				Rate:              inst.VibratoRate,
			},
		}

		ii := instrument.PCM{
			Length:        int(si.Length),
			LoopMode:      itLoopModeToLoopMode(si.Flags.LoopMode()),
			LoopBegin:     int(si.LoopStart),
			LoopEnd:       int(si.LoopStart + si.LoopLength),
			NumChannels:   1,
			Format:        instrument.SampleDataFormat8BitSigned,
			VolumeFadeout: volume.Volume(inst.VolumeFadeout) / 65536,
			Panning:       util.PanningFromIt(si.Panning),
			VolEnv: instrument.InstEnv{
				Enabled:        (inst.VolFlags & itfile.EnvelopeFlagEnabled) != 0,
				LoopEnabled:    (inst.VolFlags & itfile.EnvelopeFlagLoopEnabled) != 0,
				SustainEnabled: (inst.VolFlags & itfile.EnvelopeFlagSustainEnabled) != 0,
				LoopStart:      int(inst.VolLoopStartPoint),
				LoopEnd:        int(inst.VolLoopEndPoint),
				SustainIndex:   int(inst.VolSustainPoint),
			},
			PanEnv: instrument.InstEnv{
				Enabled:        (inst.PanFlags & itfile.EnvelopeFlagEnabled) != 0,
				LoopEnabled:    (inst.PanFlags & itfile.EnvelopeFlagLoopEnabled) != 0,
				SustainEnabled: (inst.PanFlags & itfile.EnvelopeFlagSustainEnabled) != 0,
				LoopStart:      int(inst.PanLoopStartPoint),
				LoopEnd:        int(inst.PanLoopEndPoint),
				SustainIndex:   int(inst.PanSustainPoint),
			},
		}

		if ii.VolEnv.LoopEnabled && ii.VolEnv.LoopStart > ii.VolEnv.LoopEnd {
			ii.VolEnv.LoopEnabled = false
		}

		if ii.PanEnv.LoopEnabled && ii.PanEnv.LoopStart > ii.PanEnv.LoopEnd {
			ii.PanEnv.LoopEnabled = false
		}

		if ii.VolEnv.Enabled {
			ii.VolEnv.Values = make([]instrument.EnvPoint, int(inst.VolPoints))
			for i := range ii.VolEnv.Values {
				x1 := int(inst.VolEnv[i].X)
				x2 := x1
				if i+1 < len(ii.VolEnv.Values) {
					x2 = int(inst.VolEnv[i+1].X)
				} else {
					x2 = math.MaxInt64
				}
				ii.VolEnv.Values[i] = instrument.EnvPoint{
					Ticks: x2 - x1,
					Y:     volume.Volume(uint8(inst.VolEnv[i].Y)) / 64,
				}
			}
		}

		if ii.PanEnv.Enabled {
			ii.PanEnv.Values = make([]instrument.EnvPoint, int(inst.VolPoints))
			for i := range ii.PanEnv.Values {
				x1 := int(inst.PanEnv[i].X)
				x2 := x1
				if i+1 < len(ii.PanEnv.Values) {
					x2 = int(inst.PanEnv[i+1].X)
				} else {
					x2 = math.MaxInt64
				}
				// IT stores pan envelope values in 0..64
				// So we have to do some gymnastics to remap the values
				panEnv01 := float64(uint8(inst.PanEnv[i].Y)) / 64
				panEnvVal := uint8(panEnv01 * 255)
				ii.PanEnv.Values[i] = instrument.EnvPoint{
					Ticks: x2 - x1,
					Y:     util.PanningFromIt(panEnvVal),
				}
			}
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
		stride := ii.NumChannels
		if si.Flags.Is16Bit() {
			ii.Format = instrument.SampleDataFormat16BitLESigned
			stride *= 2
		}
		ii.Length /= stride
		ii.LoopBegin /= stride
		ii.LoopEnd /= stride

		ii.Sample = si.SampleData

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

func itLoopModeToLoopMode(mode itfile.SampleLoopMode) instrument.LoopMode {
	switch mode {
	case itfile.SampleLoopModeDisabled:
		return instrument.LoopModeDisabled
	case itfile.SampleLoopModeEnabled:
		return instrument.LoopModeNormalType2
	case itfile.SampleLoopModePingPong:
		return instrument.LoopModePingPong
	default:
		return instrument.LoopModeDisabled
	}
}

func convertITInstrumentToInstrument(s itfile.IMPIIntf, linearFrequencySlides bool) ([]*instrument.Instrument, map[int][]note.Semitone, error) {
	if s == nil {
		return nil, nil, errors.New("instrument is nil")
	}

	return itInstrumentToInstrument(s, linearFrequencySlides)
}

*/

func dumpBytes(data []byte, s int) {
	r := bytes.NewReader(data)
	i := 0
	fmt.Printf("%0.4X  ", i+s)
	for {
		b, err := r.ReadByte()
		if err != nil {
			break
		}
		fmt.Printf("%0.2X", b)
		i++
		if i&15 == 0 {
			fmt.Printf("\n%0.4X  ", i+s)
		} else {
			fmt.Printf(" ")
		}
	}
	if i%15 != 0 {
		fmt.Println()
	}
}

func convertItPattern(pkt itfile.PackedPattern, channels int) (*pattern.Pattern, int, error) {
	pat := &pattern.Pattern{
		Orig: pkt,
	}

	//dumpBytes(pkt.Data, 0)

	channelMem := make([]itfile.ChannelData, channels)
	maxCh := uint8(0)
	pos := 0
	for rowNum := 0; rowNum < int(pkt.Rows); rowNum++ {
		pat.Rows = append(pat.Rows, pattern.RowData{})
		row := &pat.Rows[rowNum]
		row.Channels = make([]intf.ChannelData, channels)
	channelLoop:
		for {
			sz, chn, err := pkt.ReadChannelData(pos, channelMem)
			if err != nil {
				return nil, 0, err
			}
			//dumpBytes(pkt.Data[pos:pos+sz], pos)
			pos += sz
			if chn == nil {
				break channelLoop
			}

			channelNum := int(chn.ChannelNumber)

			cd := channel.Data{
				What:            chn.Flags,
				Note:            chn.Note,
				Instrument:      chn.Instrument,
				VolPan:          chn.VolPan,
				Effect:          chn.Command,
				EffectParameter: chn.CommandData,
			}

			row.Channels[channelNum] = &cd
			if maxCh < uint8(channelNum) {
				maxCh = uint8(channelNum)
			}
		}
	}

	return pat, int(maxCh), nil
}

func convertItFileToSong(f *itfile.File) (*layout.Song, error) {
	h, err := moduleHeaderToHeader(&f.Head)
	if err != nil {
		return nil, err
	}

	linearFrequencySlides := f.Head.Flags.IsLinearSlides()
	oldEffectMode := f.Head.Flags.IsOldEffects()

	song := layout.Song{
		Head:              *h,
		Instruments:       make(map[uint8]*instrument.Instrument),
		InstrumentNoteMap: make(map[uint8]map[note.Semitone]*instrument.Instrument),
		Patterns:          make([]pattern.Pattern, len(f.Patterns)),
		OrderList:         make([]intf.PatternIdx, int(f.Head.OrderCount)),
	}

	for i := 0; i < int(f.Head.OrderCount); i++ {
		song.OrderList[i] = intf.PatternIdx(f.OrderList[i])
	}

	if f.Head.Flags.IsUseInstruments() {
		sampNum := 0
		for instNum, inst := range f.Instruments {
			switch ii := inst.(type) {
			case *itfile.IMPIInstrumentOld:
				samples, noteMap, err := convertITInstrumentOldToInstrument(ii, f.Samples[sampNum:], linearFrequencySlides)
				if err != nil {
					return nil, err
				}

				addSamplesWithNoteMapToSong(&song, samples, noteMap, instNum)
			case *itfile.IMPIInstrument:
				samples, noteMap, err := convertITInstrumentToInstrument(ii, f.Samples[sampNum:], linearFrequencySlides)
				if err != nil {
					return nil, err
				}

				addSamplesWithNoteMapToSong(&song, samples, noteMap, instNum)
			}
		}
	}

	lastEnabledChannel := 0
	song.Patterns = make([]pattern.Pattern, len(f.Patterns))
	for patNum, pkt := range f.Patterns {
		pattern, maxCh, err := convertItPattern(pkt, len(f.Head.ChannelVol))
		if err != nil {
			return nil, err
		}
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
			InitialVolume:  volume.Volume(f.Head.ChannelVol[chNum].Value()),
			InitialPanning: util.PanningFromIt(f.Head.ChannelPan[chNum]),
			Memory: channel.Memory{
				LinearFreqSlides: linearFrequencySlides,
				OldEffectMode:    oldEffectMode,
			},
		}

		channels[chNum] = cs
	}

	song.ChannelSettings = channels

	return &song, nil
}

func addSamplesWithNoteMapToSong(song *layout.Song, samples []*instrument.Instrument, noteMap map[int][]note.Semitone, instNum int) {
	for _, sample := range samples {
		if sample == nil {
			continue
		}
		id := channel.SampleID{
			InstID: uint8(instNum + 1),
		}
		sample.ID = id
		song.Instruments[id.InstID] = sample
	}
	for i, sts := range noteMap {
		sample := samples[i]
		id, ok := sample.ID.(channel.SampleID)
		if !ok {
			continue
		}
		inm, ok := song.InstrumentNoteMap[id.InstID]
		if !ok {
			inm = make(map[note.Semitone]*instrument.Instrument)
			song.InstrumentNoteMap[id.InstID] = inm
		}
		for _, st := range sts {
			inm[st] = samples[i]
		}
	}
}

func readIT(filename string) (*layout.Song, error) {
	buffer, err := formatutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	s, err := itfile.Read(buffer)
	if err != nil {
		return nil, err
	}

	return convertItFileToSong(s)
}
