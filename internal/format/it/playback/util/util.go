package util

import (
	"math"

	itfile "github.com/gotracker/goaudiofile/music/tracked/it"
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/note"
)

const (
	// DefaultC2Spd is the default C2SPD for IT samples
	DefaultC2Spd = 8363

	floatDefaultC2Spd = float32(DefaultC2Spd)
	// C5Period is the sampler (Amiga-style) period of the C-5 note
	C5Period = float32(1712)

	// ITBaseClock is the base clock speed of IT files
	ITBaseClock = floatDefaultC2Spd * C5Period
)

var (
	// DefaultVolume is the default volume value for most everything in it format
	DefaultVolume = VolumeFromIt(0x40)

	// DefaultMixingVolume is the default mixing volume
	DefaultMixingVolume = volume.Volume(0x30) / 0x80

	// DefaultPanningLeft is the default panning value for left channels
	DefaultPanningLeft = PanningFromIt(0x30)
	// DefaultPanning is the default panning value for unconfigured channels
	DefaultPanning = PanningFromIt(0x80)
	// DefaultPanningRight is the default panning value for right channels
	DefaultPanningRight = PanningFromIt(0xC0)
)

var semitonePeriodTable = [...]float32{27392, 25856, 24384, 23040, 21696, 20480, 19328, 18240, 17216, 16256, 15360, 14496}

// CalcSemitonePeriod calculates the semitone period for it notes
func CalcSemitonePeriod(semi note.Semitone, ft note.Finetune, c2spd note.C2SPD, linearFreqSlides bool) note.Period {
	if semi == note.UnchangedSemitone {
		panic("how?")
	}
	if linearFreqSlides {
		nft := int(semi)*64 + int(ft)
		return &LinearPeriod{
			// NOTE: not sure why the magic downshift a whole octave,
			// but it makes all the calculations work, so here we are.
			Finetune: note.Finetune(nft),
			C2Spd:    c2spd,
		}
	}

	key := int(semi.Key())
	octave := uint32(semi.Octave())

	if key >= len(semitonePeriodTable) {
		return nil
	}

	if c2spd == 0 {
		c2spd = note.C2SPD(DefaultC2Spd)
	}

	if ft != 0 {
		c2spd = CalcFinetuneC2Spd(c2spd, ft, linearFreqSlides)
	}

	period := (AmigaPeriod(floatDefaultC2Spd*semitonePeriodTable[key]) / AmigaPeriod(uint32(c2spd)<<octave))
	period = period.AddInteger(0)
	return &period
}

// CalcFinetuneC2Spd calculates a new C2SPD after a finetune adjustment
func CalcFinetuneC2Spd(c2spd note.C2SPD, finetune note.Finetune, linearFreqSlides bool) note.C2SPD {
	if finetune == 0 {
		return c2spd
	}

	nft := (5*12)*64 + int(finetune)
	period := CalcSemitonePeriod(note.Semitone(nft/64), note.Finetune(nft%64), c2spd, linearFreqSlides)
	return note.C2SPD(period.GetFrequency())
}

// VolumeFromIt converts an it volume to a player volume
func VolumeFromIt(vol itfile.Volume) volume.Volume {
	return volume.Volume(vol.Value())
}

// VolumeFromVolPan converts an it volume-pan to a player volume
func VolumeFromVolPan(vp uint8) volume.Volume {
	switch {
	case vp >= 0 && vp <= 64:
		return volume.Volume(vp) / 64
	default:
		return volume.VolumeUseInstVol
	}
}

// VolumeToIt converts a player volume to an it volume
func VolumeToIt(v volume.Volume) itfile.Volume {
	switch {
	case v == volume.VolumeUseInstVol:
		return 0
	default:
		return itfile.Volume(v * 64.0)
	}
}

// PanningFromIt returns a radian panning position from an it panning value
func PanningFromIt(pos itfile.PanValue) panning.Position {
	if pos.IsDisabled() {
		return panning.CenterAhead
	}
	return panning.MakeStereoPosition(pos.Value(), 0, 1)
}

// PanningToIt returns the it panning value for a radian panning position
func PanningToIt(pan panning.Position) itfile.PanValue {
	p := panning.FromStereoPosition(pan, 0, 1)
	return itfile.PanValue(p * 64)
}

// NoteFromItNote converts an it file note into a player note
func NoteFromItNote(in itfile.Note) note.Note {
	switch {
	case in.IsNoteOff():
		return note.ReleaseNote
	case in.IsNoteCut():
		return note.StopNote
	case in.IsNoteFade(): // not really invalid, but...
		return note.InvalidNote
	}

	an := uint8(in)
	s := note.Semitone(an)
	return note.NewNote(s)
}

// FrequencyFromSemitone returns the frequency from the semitone (and c2spd)
func FrequencyFromSemitone(semitone note.Semitone, c2spd note.C2SPD, linearFreqSlides bool) float32 {
	period := CalcSemitonePeriod(semitone, 0, c2spd, linearFreqSlides)
	return float32(period.GetFrequency())
}

// ToAmigaPeriod calculates an amiga period for a linear finetune period
func ToAmigaPeriod(finetunes note.Finetune, c2spd note.C2SPD) AmigaPeriod {
	if finetunes < 0 {
		finetunes = 0
	}
	pow := math.Pow(2, float64(finetunes)/768)
	linFreq := float64(c2spd) * pow / float64(DefaultC2Spd)

	period := AmigaPeriod(float64(semitonePeriodTable[0]) / linFreq)
	return period
}

// ToLinearPeriod returns the linear frequency period for a given period
func ToLinearPeriod(p note.Period) *LinearPeriod {
	switch pp := p.(type) {
	case *LinearPeriod:
		return pp
	case *AmigaPeriod:
		linFreq := float64(semitonePeriodTable[0]) / float64(*pp)

		fts := note.Finetune(768 * math.Log2(linFreq))

		lp := LinearPeriod{
			Finetune: fts,
			C2Spd:    DefaultC2Spd,
		}
		return &lp
	}
	return nil
}
