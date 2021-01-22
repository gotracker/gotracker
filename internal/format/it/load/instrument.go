package load

import (
	itfile "github.com/gotracker/goaudiofile/music/tracked/it"
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/format/it/playback/util"
	"gotracker/internal/instrument"
	"gotracker/internal/player/note"
)

func convertITInstrumentOldToInstrument(inst *itfile.IMPIInstrumentOld, sampData []itfile.FullSample, linearFrequencySlides bool) ([]*instrument.Instrument, map[int][]note.Semitone, error) {
	samples := make([]*instrument.Instrument, 0)
	noteKeyboard := make(map[int][]note.Semitone)
	for i := 0; i < int(inst.SampleCount); i++ {
		si := &sampData[i]

		id := instrument.PCM{
			Sample: si.Data,
			Length: int(si.Header.Length),
			Loop: instrument.LoopInfo{
				Mode:  instrument.LoopModeDisabled,
				Begin: int(si.Header.LoopBegin),
				End:   int(si.Header.LoopEnd),
			},
			SustainLoop: instrument.LoopInfo{
				Mode:  instrument.LoopModeDisabled,
				Begin: int(si.Header.SustainLoopBegin),
				End:   int(si.Header.SustainLoopEnd),
			},
			NumChannels:   1,
			Format:        instrument.SampleDataFormat8BitUnsigned,
			Panning:       panning.CenterAhead,
			VolumeFadeout: volume.Volume(1),
		}

		if si.Header.Flags.IsLoopEnabled() {
			if si.Header.Flags.IsLoopPingPong() {
				id.Loop.Mode = instrument.LoopModePingPong
			} else {
				id.Loop.Mode = instrument.LoopModeNormalType2
			}
		}

		if si.Header.Flags.IsSustainLoopEnabled() {
			if si.Header.Flags.IsSustainLoopPingPong() {
				id.Loop.Mode = instrument.LoopModePingPong
			} else {
				id.Loop.Mode = instrument.LoopModeNormalType2
			}
		}

		if si.Header.Flags.IsStereo() {
			id.NumChannels = 2
		}

		is16Bit := si.Header.Flags.Is16Bit()
		isSigned := si.Header.ConvertFlags.IsSignedSamples()
		isBigEndian := si.Header.ConvertFlags.IsBigEndian()
		id.Format = getSampleFormat(is16Bit, isSigned, isBigEndian)

		ii := instrument.Instrument{
			Filename: si.Header.GetFilename(),
			Name:     si.Header.GetName(),
			Inst:     &id,
			C2Spd:    note.C2SPD(si.Header.C5Speed),
			AutoVibrato: instrument.AutoVibrato{
				Enabled:           (si.Header.VibratoDepth != 0 && si.Header.VibratoSpeed != 0 && si.Header.VibratoSweep != 0),
				Sweep:             0,
				WaveformSelection: si.Header.VibratoType,
				Depth:             si.Header.VibratoDepth,
				Rate:              si.Header.VibratoSpeed,
			},
			Volume: volume.Volume(si.Header.Volume.Value()),
		}
		if si.Header.VibratoSweep != 0 {
			ii.AutoVibrato.Sweep = uint8(int(si.Header.VibratoDepth) * 256 / int(si.Header.VibratoSweep))
		}
		if !si.Header.DefaultPan.IsDisabled() {
			id.Panning = panning.MakeStereoPosition(si.Header.DefaultPan.Value(), 0, 1)
		}

		samples = append(samples, &ii)
	}

	for _, ns := range inst.NoteSampleKeyboard {
		s := int(ns.Sample)
		if s == 0 {
			continue
		}
		si := int(ns.Sample) - 1
		noteKeyboard[si] = append(noteKeyboard[si], util.NoteFromItNote(ns.Note).Semitone())
	}

	return samples, noteKeyboard, nil
}

func convertITInstrumentToInstrument(inst *itfile.IMPIInstrument, sampData []itfile.FullSample, linearFrequencySlides bool) ([]*instrument.Instrument, map[int][]note.Semitone, error) {
	samples := make([]*instrument.Instrument, 0)
	noteKeyboard := make(map[int][]note.Semitone)
	for i := 0; i < int(inst.SampleCount); i++ {
		si := &sampData[i]

		id := instrument.PCM{
			Sample: si.Data,
			Length: int(si.Header.Length),
			Loop: instrument.LoopInfo{
				Mode:  instrument.LoopModeDisabled,
				Begin: int(si.Header.LoopBegin),
				End:   int(si.Header.LoopEnd),
			},
			SustainLoop: instrument.LoopInfo{
				Mode:  instrument.LoopModeDisabled,
				Begin: int(si.Header.SustainLoopBegin),
				End:   int(si.Header.SustainLoopEnd),
			},
			NumChannels:   1,
			Format:        instrument.SampleDataFormat8BitUnsigned,
			Panning:       panning.CenterAhead,
			VolumeFadeout: volume.Volume(1),
		}

		if si.Header.Flags.IsLoopEnabled() {
			if si.Header.Flags.IsLoopPingPong() {
				id.Loop.Mode = instrument.LoopModePingPong
			} else {
				id.Loop.Mode = instrument.LoopModeNormalType2
			}
		}

		if si.Header.Flags.IsSustainLoopEnabled() {
			if si.Header.Flags.IsSustainLoopPingPong() {
				id.Loop.Mode = instrument.LoopModePingPong
			} else {
				id.Loop.Mode = instrument.LoopModeNormalType2
			}
		}

		if si.Header.Flags.IsStereo() {
			id.NumChannels = 2
		}

		is16Bit := si.Header.Flags.Is16Bit()
		isSigned := si.Header.ConvertFlags.IsSignedSamples()
		isBigEndian := si.Header.ConvertFlags.IsBigEndian()
		id.Format = getSampleFormat(is16Bit, isSigned, isBigEndian)

		ii := instrument.Instrument{
			Filename: si.Header.GetFilename(),
			Name:     si.Header.GetName(),
			Inst:     &id,
			C2Spd:    note.C2SPD(si.Header.C5Speed),
			AutoVibrato: instrument.AutoVibrato{
				Enabled:           (si.Header.VibratoDepth != 0 && si.Header.VibratoSpeed != 0 && si.Header.VibratoSweep != 0),
				Sweep:             0,
				WaveformSelection: si.Header.VibratoType,
				Depth:             si.Header.VibratoDepth,
				Rate:              si.Header.VibratoSpeed,
			},
			Volume: volume.Volume(si.Header.Volume.Value()),
		}
		if si.Header.VibratoSweep != 0 {
			ii.AutoVibrato.Sweep = uint8(int(si.Header.VibratoDepth) * 256 / int(si.Header.VibratoSweep))
		}
		if !si.Header.DefaultPan.IsDisabled() {
			id.Panning = panning.MakeStereoPosition(si.Header.DefaultPan.Value(), 0, 1)
		}

		samples = append(samples, &ii)
	}

	for _, ns := range inst.NoteSampleKeyboard {
		s := int(ns.Sample)
		if s == 0 {
			continue
		}
		si := int(ns.Sample) - 1
		n := util.NoteFromItNote(ns.Note)
		st := n.Semitone()
		noteKeyboard[si] = append(noteKeyboard[si], st)
	}

	return samples, noteKeyboard, nil
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
