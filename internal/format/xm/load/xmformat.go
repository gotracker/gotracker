package load

import (
	"errors"
	"math"

	xmfile "github.com/gotracker/goaudiofile/music/tracked/xm"
	"github.com/gotracker/gomixing/volume"

	formatutil "gotracker/internal/format/internal/util"
	"gotracker/internal/format/xm/layout"
	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/instrument"
	"gotracker/internal/oscillator"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/pattern"
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

func xmInstrumentToInstrument(inst *xmfile.InstrumentHeader, linearFrequencySlides bool) ([]*instrument.Instrument, map[int][]note.Semitone, error) {
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
			Volume:             util.VolumeFromXm(0x10 + v),
			RelativeNoteNumber: si.RelativeNoteNumber,
			AutoVibrato: instrument.AutoVibrato{
				Enabled:           (inst.VibratoDepth != 0 && inst.VibratoRate != 0),
				Sweep:             inst.VibratoSweep, // NOTE: for IT support, this needs to be calculated as (Depth * 256 / VibratoSweep) ticks
				WaveformSelection: inst.VibratoType,
				Depth:             inst.VibratoDepth,
				Rate:              inst.VibratoRate,
				Factory:           oscillator.NewProtrackerOscillator,
			},
		}

		ii := instrument.PCM{
			Length: int(si.Length),
			Loop: instrument.LoopInfo{
				Mode: instrument.LoopModeDisabled,
			},
			SustainLoop: instrument.LoopInfo{
				Mode:  xmLoopModeToLoopMode(si.Flags.LoopMode()),
				Begin: int(si.LoopStart),
				End:   int(si.LoopStart + si.LoopLength),
			},
			NumChannels:  1,
			Format:       instrument.SampleDataFormat8BitSigned,
			MixingVolume: volume.Volume(1),
			FadeOut: instrument.FadeoutSettings{
				Mode:   instrument.FadeoutModeOnlyIfVolEnvActive,
				Amount: volume.Volume(inst.VolumeFadeout) / 65536,
			},
			Panning: util.PanningFromXm(si.Panning),
			VolEnv: instrument.InstEnv{
				Enabled:          (inst.VolFlags & xmfile.EnvelopeFlagEnabled) != 0,
				LoopEnabled:      (inst.VolFlags & xmfile.EnvelopeFlagLoopEnabled) != 0,
				SustainEnabled:   (inst.VolFlags & xmfile.EnvelopeFlagSustainEnabled) != 0,
				LoopStart:        int(inst.VolLoopStartPoint),
				LoopEnd:          int(inst.VolLoopEndPoint),
				SustainLoopStart: int(inst.VolSustainPoint),
				SustainLoopEnd:   int(inst.VolSustainPoint) + 1,
			},
			PanEnv: instrument.InstEnv{
				Enabled:          (inst.PanFlags & xmfile.EnvelopeFlagEnabled) != 0,
				LoopEnabled:      (inst.PanFlags & xmfile.EnvelopeFlagLoopEnabled) != 0,
				SustainEnabled:   (inst.PanFlags & xmfile.EnvelopeFlagSustainEnabled) != 0,
				LoopStart:        int(inst.PanLoopStartPoint),
				LoopEnd:          int(inst.PanLoopEndPoint),
				SustainLoopStart: int(inst.PanSustainPoint),
				SustainLoopEnd:   int(inst.PanSustainPoint) + 1,
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
				// XM stores pan envelope values in 0..64
				// So we have to do some gymnastics to remap the values
				panEnv01 := float64(uint8(inst.PanEnv[i].Y)) / 64
				panEnvVal := uint8(panEnv01 * 255)
				ii.PanEnv.Values[i] = instrument.EnvPoint{
					Ticks: x2 - x1,
					Y:     util.PanningFromXm(panEnvVal),
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
		ii.SustainLoop.Begin /= stride
		ii.SustainLoop.End /= stride

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

func xmLoopModeToLoopMode(mode xmfile.SampleLoopMode) instrument.LoopMode {
	switch mode {
	case xmfile.SampleLoopModeDisabled:
		return instrument.LoopModeDisabled
	case xmfile.SampleLoopModeEnabled:
		return instrument.LoopModeNormalType2
	case xmfile.SampleLoopModePingPong:
		return instrument.LoopModePingPong
	default:
		return instrument.LoopModeDisabled
	}
}

func convertXMInstrumentToInstrument(s *xmfile.InstrumentHeader, linearFrequencySlides bool) ([]*instrument.Instrument, map[int][]note.Semitone, error) {
	if s == nil {
		return nil, nil, errors.New("instrument is nil")
	}

	return xmInstrumentToInstrument(s, linearFrequencySlides)
}

func convertXmPattern(pkt xmfile.Pattern) (*pattern.Pattern, int) {
	pat := &pattern.Pattern{
		Orig: pkt,
	}

	maxCh := uint8(0)
	for rowNum, drow := range pkt.Data {
		pat.Rows = append(pat.Rows, pattern.RowData{})
		row := &pat.Rows[rowNum]
		row.Channels = make([]intf.ChannelData, len(drow))
		for channelNum, chn := range drow {
			cd := channel.Data{
				What:            chn.Flags,
				Note:            chn.Note,
				Instrument:      chn.Instrument,
				Volume:          chn.Volume,
				Effect:          chn.Effect,
				EffectParameter: chn.EffectParameter,
			}
			row.Channels[channelNum] = &cd
			if maxCh < uint8(channelNum) {
				maxCh = uint8(channelNum)
			}
		}
	}

	return pat, int(maxCh)
}

func convertXmFileToSong(f *xmfile.File) (*layout.Song, error) {
	h, err := moduleHeaderToHeader(&f.Head)
	if err != nil {
		return nil, err
	}

	linearFrequencySlides := f.Head.Flags.IsLinearSlides()

	song := layout.Song{
		Head:              *h,
		Instruments:       make(map[uint8]*instrument.Instrument),
		InstrumentNoteMap: make(map[uint8]map[note.Semitone]*instrument.Instrument),
		Patterns:          make([]pattern.Pattern, len(f.Patterns)),
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

	lastEnabledChannel := 0
	song.Patterns = make([]pattern.Pattern, len(f.Patterns))
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

		cs.Memory.ResetOscillators()

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
