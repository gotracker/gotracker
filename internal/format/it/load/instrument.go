package load

import (
	"math"

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
			VolumeFadeout: volume.Volume(inst.Fadeout) / (64 * 512),
			VolEnv: instrument.InstEnv{
				Enabled:          (inst.Flags & itfile.IMPIOldFlagUseVolumeEnvelope) != 0,
				LoopEnabled:      (inst.Flags & itfile.IMPIOldFlagUseVolumeLoop) != 0,
				SustainEnabled:   (inst.Flags & itfile.IMPIOldFlagUseSustainVolumeLoop) != 0,
				LoopStart:        int(inst.VolumeLoopStart),
				LoopEnd:          int(inst.VolumeLoopEnd),
				SustainLoopStart: int(inst.SustainLoopStart),
				SustainLoopEnd:   int(inst.SustainLoopEnd),
				Values:           make([]instrument.EnvPoint, 0),
			},
		}

		for i := range inst.VolumeEnvelope {
			out := instrument.EnvPoint{}
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
			id.VolEnv.Values = append(id.VolEnv.Values, out)
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
			VolumeFadeout: volume.Volume(inst.Fadeout) / (128 * 1024),
			VolEnv: instrument.InstEnv{
				Enabled:          (inst.VolumeEnvelope.Flags & itfile.EnvelopeFlagEnvelopeOn) != 0,
				LoopEnabled:      (inst.VolumeEnvelope.Flags & itfile.EnvelopeFlagLoopOn) != 0,
				SustainEnabled:   (inst.VolumeEnvelope.Flags & itfile.EnvelopeFlagSustainLoopOn) != 0,
				LoopStart:        int(inst.VolumeEnvelope.LoopBegin),
				LoopEnd:          int(inst.VolumeEnvelope.LoopEnd),
				SustainLoopStart: int(inst.VolumeEnvelope.SustainLoopBegin),
				SustainLoopEnd:   int(inst.VolumeEnvelope.SustainLoopEnd),
				Values:           make([]instrument.EnvPoint, int(inst.VolumeEnvelope.Count)),
			},
			PanEnv: instrument.InstEnv{
				Enabled:          (inst.PanningEnvelope.Flags & itfile.EnvelopeFlagEnvelopeOn) != 0,
				LoopEnabled:      (inst.PanningEnvelope.Flags & itfile.EnvelopeFlagLoopOn) != 0,
				SustainEnabled:   (inst.PanningEnvelope.Flags & itfile.EnvelopeFlagSustainLoopOn) != 0,
				LoopStart:        int(inst.PanningEnvelope.LoopBegin),
				LoopEnd:          int(inst.PanningEnvelope.LoopEnd),
				SustainLoopStart: int(inst.PanningEnvelope.SustainLoopBegin),
				SustainLoopEnd:   int(inst.PanningEnvelope.SustainLoopEnd),
				Values:           make([]instrument.EnvPoint, int(inst.PanningEnvelope.Count)),
			},
			PitchEnv: instrument.InstEnv{
				Enabled:          (inst.PitchEnvelope.Flags & itfile.EnvelopeFlagEnvelopeOn) != 0,
				LoopEnabled:      (inst.PitchEnvelope.Flags & itfile.EnvelopeFlagLoopOn) != 0,
				SustainEnabled:   (inst.PitchEnvelope.Flags & itfile.EnvelopeFlagSustainLoopOn) != 0,
				LoopStart:        int(inst.PitchEnvelope.LoopBegin),
				LoopEnd:          int(inst.PitchEnvelope.LoopEnd),
				SustainLoopStart: int(inst.PitchEnvelope.SustainLoopBegin),
				SustainLoopEnd:   int(inst.PitchEnvelope.SustainLoopEnd),
				Values:           make([]instrument.EnvPoint, int(inst.PitchEnvelope.Count)),
			},
		}

		for i := range id.VolEnv.Values {
			out := &id.VolEnv.Values[i]
			in1 := inst.VolumeEnvelope.NodePoints[i]
			vol := volume.Volume(uint8(in1.Y)) / 64
			if vol > 1 {
				// NOTE: there might be an incoming Y value == 0xFF, which really
				// means "end of envelope" and should not mean "full volume",
				// but we can cheat a little here and probably get away with it...
				vol = 1
			}
			out.Y = vol
			if i+1 < len(id.VolEnv.Values) {
				in2 := inst.VolumeEnvelope.NodePoints[i+1]
				out.Ticks = int(in2.Tick) - int(in1.Tick)
			} else {
				out.Ticks = math.MaxInt64
			}
		}

		for i := range id.PanEnv.Values {
			out := &id.PanEnv.Values[i]
			in1 := inst.PanningEnvelope.NodePoints[i]
			out.Y = util.PanningFromIt(itfile.PanValue(in1.Y))
			if i+1 < len(id.PanEnv.Values) {
				in2 := inst.PanningEnvelope.NodePoints[i+1]
				out.Ticks = int(in2.Tick) - int(in1.Tick)
			} else {
				out.Ticks = math.MaxInt64
			}
		}

		for i := range id.PitchEnv.Values {
			out := &id.PitchEnv.Values[i]
			in1 := inst.PitchEnvelope.NodePoints[i]
			out.Y = note.PeriodDelta(in1.Y)
			if i+1 < len(id.PitchEnv.Values) {
				in2 := inst.PitchEnvelope.NodePoints[i+1]
				out.Ticks = int(in2.Tick) - int(in1.Tick)
			} else {
				out.Ticks = math.MaxInt64
			}
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
