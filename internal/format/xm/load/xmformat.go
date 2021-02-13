package load

import (
	"errors"
	"math"

	xmfile "github.com/gotracker/goaudiofile/music/tracked/xm"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/envelope"
	formatutil "gotracker/internal/format/internal/util"
	"gotracker/internal/format/xm/layout"
	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/format/xm/playback/util"
	"gotracker/internal/instrument"
	"gotracker/internal/loop"
	"gotracker/internal/oscillator"
	"gotracker/internal/pcm"
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
		v := util.VolumeXM(si.Volume)
		if v >= 0x40 {
			v = 0x40
		}
		sample := instrument.Instrument{
			Filename:           si.GetName(),
			Name:               inst.GetName(),
			C2Spd:              note.C2SPD(0), // uses si.Finetune, below
			Volume:             v.Volume(),
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

		instLen := int(si.Length)
		numChannels := 1
		format := pcm.SampleDataFormat8BitSigned

		sustainMode := xmLoopModeToLoopMode(si.Flags.LoopMode())
		sustainSettings := loop.Settings{
			Begin: int(si.LoopStart),
			End:   int(si.LoopStart + si.LoopLength),
		}

		volEnvLoopMode := loop.ModeDisabled
		volEnvLoopSettings := loop.Settings{
			Begin: int(inst.VolLoopStartPoint),
			End:   int(inst.VolLoopEndPoint),
		}
		volEnvSustainMode := loop.ModeDisabled
		volEnvSustainSettings := loop.Settings{
			Begin: int(inst.VolSustainPoint),
			End:   int(inst.VolSustainPoint),
		}

		panEnvLoopMode := loop.ModeDisabled
		panEnvLoopSettings := loop.Settings{
			Begin: int(inst.PanLoopStartPoint),
			End:   int(inst.PanLoopEndPoint),
		}
		panEnvSustainMode := loop.ModeDisabled
		panEnvSustainSettings := loop.Settings{
			Begin: int(inst.PanSustainPoint),
			End:   int(inst.PanSustainPoint),
		}

		ii := instrument.PCM{
			Loop:         &loop.Disabled{},
			MixingVolume: volume.Volume(1),
			FadeOut: instrument.FadeoutSettings{
				Mode:   instrument.FadeoutModeOnlyIfVolEnvActive,
				Amount: volume.Volume(inst.VolumeFadeout) / 65536,
			},
			Panning: util.PanningFromXm(si.Panning),
			VolEnv: envelope.Envelope{
				Enabled: (inst.VolFlags & xmfile.EnvelopeFlagEnabled) != 0,
			},
			PanEnv: envelope.Envelope{
				Enabled: (inst.PanFlags & xmfile.EnvelopeFlagEnabled) != 0,
			},
		}

		if ii.VolEnv.Enabled && ii.VolEnv.Loop.Length() >= 0 {
			if enabled := (inst.VolFlags & xmfile.EnvelopeFlagLoopEnabled) != 0; enabled {
				volEnvLoopMode = loop.ModeNormal
			}
			if enabled := (inst.VolFlags & xmfile.EnvelopeFlagSustainEnabled) != 0; enabled {
				volEnvSustainMode = loop.ModeNormal
			}

			ii.VolEnv.Values = make([]envelope.EnvPoint, int(inst.VolPoints))
			for i := range ii.VolEnv.Values {
				x1 := int(inst.VolEnv[i].X)
				y1 := uint8(inst.VolEnv[i].Y)
				x2 := x1
				if i+1 < len(ii.VolEnv.Values) {
					x2 = int(inst.VolEnv[i+1].X)
				} else {
					x2 = math.MaxInt64
				}
				ii.VolEnv.Values[i] = &envelope.VolumePoint{
					Ticks: x2 - x1,
					Y:     util.VolumeXM(y1).Volume(),
				}
			}
		}

		if ii.PanEnv.Enabled && ii.PanEnv.Loop.Length() >= 0 {
			if enabled := (inst.PanFlags & xmfile.EnvelopeFlagLoopEnabled) != 0; enabled {
				panEnvLoopMode = loop.ModeNormal
			}
			if enabled := (inst.PanFlags & xmfile.EnvelopeFlagSustainEnabled) != 0; enabled {
				panEnvSustainMode = loop.ModeNormal
			}

			ii.PanEnv.Values = make([]envelope.EnvPoint, int(inst.VolPoints))
			for i := range ii.PanEnv.Values {
				x1 := int(inst.PanEnv[i].X)
				// XM stores pan envelope values in 0..64
				// So we have to do some gymnastics to remap the values
				panEnv01 := float64(uint8(inst.PanEnv[i].Y)) / 64
				y1 := uint8(panEnv01 * 255)
				x2 := x1
				if i+1 < len(ii.PanEnv.Values) {
					x2 = int(inst.PanEnv[i+1].X)
				} else {
					x2 = math.MaxInt64
				}
				ii.PanEnv.Values[i] = &envelope.PanPoint{
					Ticks: x2 - x1,
					Y:     util.PanningFromXm(y1),
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
			numChannels = 2
		}
		stride := numChannels
		if si.Flags.Is16Bit() {
			format = pcm.SampleDataFormat16BitLESigned
			stride *= 2
		}
		instLen /= stride
		sustainSettings.Begin /= stride
		sustainSettings.End /= stride

		ii.SustainLoop = loop.NewLoop(sustainMode, sustainSettings)
		ii.VolEnv.Loop = loop.NewLoop(volEnvLoopMode, volEnvLoopSettings)
		ii.VolEnv.Sustain = loop.NewLoop(volEnvSustainMode, volEnvSustainSettings)
		ii.PanEnv.Loop = loop.NewLoop(panEnvLoopMode, panEnvLoopSettings)
		ii.PanEnv.Sustain = loop.NewLoop(panEnvSustainMode, panEnvSustainSettings)

		ii.Sample = pcm.NewSample(si.SampleData, instLen, numChannels, format)

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

func xmLoopModeToLoopMode(mode xmfile.SampleLoopMode) loop.Mode {
	switch mode {
	case xmfile.SampleLoopModeDisabled:
		return loop.ModeDisabled
	case xmfile.SampleLoopModeEnabled:
		return loop.ModeNormal
	case xmfile.SampleLoopModePingPong:
		return loop.ModePingPong
	default:
		return loop.ModeDisabled
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
				Volume:          util.VolEffect(chn.Volume),
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
