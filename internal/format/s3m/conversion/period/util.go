package period

import (
	s3mfile "github.com/gotracker/goaudiofile/music/tracked/s3m"
	"github.com/gotracker/gotracker/internal/song/note"
	"github.com/gotracker/voice/period"
)

const (
	floatDefaultC2Spd = float32(DefaultC2Spd)
	c2Period          = 1712

	// DefaultC2Spd is the default C2SPD for S3M samples
	DefaultC2Spd = period.Frequency(s3mfile.DefaultC2Spd)

	// S3MBaseClock is the base clock speed of S3M files
	S3MBaseClock period.Frequency = DefaultC2Spd * c2Period

	notesPerOctave     = 12
	semitonesPerNote   = 64
	semitonesPerOctave = notesPerOctave * semitonesPerNote
)

var semitonePeriodTable = [...]float32{27392, 25856, 24384, 23040, 21696, 20480, 19328, 18240, 17216, 16256, 15360, 14496}

// CalcSemitonePeriod calculates the semitone period for it notes
func CalcSemitonePeriod(semi note.Semitone, ft note.Finetune, c2spd note.C2SPD) note.Period {
	if semi == note.UnchangedSemitone {
		panic("how?")
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
		c2spd = CalcFinetuneC2Spd(c2spd, ft)
	}

	p := (Amiga(floatDefaultC2Spd*semitonePeriodTable[key]) / Amiga(uint32(c2spd)<<octave))
	p = p.AddInteger(0)
	return p
}

// CalcFinetuneC2Spd calculates a new C2SPD after a finetune adjustment
func CalcFinetuneC2Spd(c2spd note.C2SPD, finetune note.Finetune) note.C2SPD {
	if finetune == 0 {
		return c2spd
	}

	nft := 5*semitonesPerOctave + int(finetune)
	p := CalcSemitonePeriod(note.Semitone(nft/semitonesPerNote), note.Finetune(nft%semitonesPerNote), c2spd)
	return note.C2SPD(p.GetFrequency())
}

// FrequencyFromSemitone returns the frequency from the semitone (and c2spd)
func FrequencyFromSemitone(semitone note.Semitone, c2spd note.C2SPD) float32 {
	p := CalcSemitonePeriod(semitone, 0, c2spd)
	return float32(p.GetFrequency())
}
