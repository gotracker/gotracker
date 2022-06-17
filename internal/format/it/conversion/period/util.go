package period

import (
	"github.com/gotracker/gotracker/internal/song/note"
)

const (
	// DefaultC2Spd is the default C2SPD for IT samples
	DefaultC2Spd = 8363

	floatDefaultC2Spd = float32(DefaultC2Spd)
	// C5Period is the sampler (Amiga-style) period of the C-5 note
	C5Period = float32(428)

	// ITBaseClock is the base clock speed of IT files
	ITBaseClock = floatDefaultC2Spd * C5Period

	notesPerOctave     = 12
	semitonesPerNote   = 64
	semitonesPerOctave = notesPerOctave * semitonesPerNote
)

var semitonePeriodTable = [...]float32{27392, 25856, 24384, 23040, 21696, 20480, 19328, 18240, 17216, 16256, 15360, 14496}

// CalcSemitonePeriod calculates the semitone period for it notes
func CalcSemitonePeriod(semi note.Semitone, ft note.Finetune, c2spd note.C2SPD, linearFreqSlides bool) note.Period {
	if semi == note.UnchangedSemitone {
		panic("how?")
	}
	if linearFreqSlides {
		nft := int(semi)*64 + int(ft)
		return &Linear{
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

	period := (Amiga(floatDefaultC2Spd*semitonePeriodTable[key]) / Amiga(uint32(c2spd)<<octave))
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

// FrequencyFromSemitone returns the frequency from the semitone (and c2spd)
func FrequencyFromSemitone(semitone note.Semitone, c2spd note.C2SPD, linearFreqSlides bool) float32 {
	period := CalcSemitonePeriod(semitone, 0, c2spd, linearFreqSlides)
	return float32(period.GetFrequency())
}
