package util

import (
	"math"

	s3mfile "github.com/gotracker/goaudiofile/music/tracked/s3m"
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/note"
)

const (
	floatDefaultC2Spd = float32(s3mfile.DefaultC2Spd)
	c2Period          = float32(1712)

	// S3MBaseClock is the base clock speed of S3M files
	S3MBaseClock = floatDefaultC2Spd * c2Period
)

var (
	// DefaultVolume is the default volume value for most everything in S3M format
	DefaultVolume = VolumeFromS3M(s3mfile.DefaultVolume)

	// DefaultPanningLeft is the default panning value for left channels
	DefaultPanningLeft = PanningFromS3M(0x03)
	// DefaultPanning is the default panning value for unconfigured channels
	DefaultPanning = PanningFromS3M(0x08)
	// DefaultPanningRight is the default panning value for right channels
	DefaultPanningRight = PanningFromS3M(0x0C)
)

// AmigaPeriod defines a sampler period that follows the Amiga-style approach of note
// definition. Useful in calculating resampling.
type AmigaPeriod float32

// AddInteger truncates the current period to an integer and adds the delta integer in
// then returns the resulting period
func (p *AmigaPeriod) AddInteger(delta int) AmigaPeriod {
	period := AmigaPeriod(int(*p) + delta)
	return period
}

// Add adds the current period to a delta value then returns the resulting period
func (p *AmigaPeriod) Add(delta note.Period) note.Period {
	period := AmigaPeriod(*p)
	if d, ok := delta.(*AmigaPeriod); ok {
		period += *d
	}
	return &period
}

// ToAmigaPeriod returns an Amiga-style period
func (p *AmigaPeriod) ToAmigaPeriod() AmigaPeriod {
	return *p
}

// Compare returns:
//  -1 if the current period is higher frequency than the `rhs` period
//  0 if the current period is equal in frequency to the `rhs` period
//  1 if the current period is lower frequency than the `rhs` period
func (p *AmigaPeriod) Compare(rhs note.Period) int {
	right := AmigaPeriod(0)
	if r, ok := rhs.(*AmigaPeriod); ok {
		right = *r
	}

	switch {
	case *p > right:
		return -1
	case *p < right:
		return 1
	default:
		return 0
	}
}

// Lerp linear-interpolates the current period with the `rhs` period
func (p *AmigaPeriod) Lerp(t float64, rhs note.Period) note.Period {
	right := AmigaPeriod(0)
	if r, ok := rhs.(*AmigaPeriod); ok {
		right = *r
	}

	period := *p
	period += AmigaPeriod(t * (float64(right) - float64(period)))
	return &period
}

// GetSamplerAdd returns the number of samples to advance an instrument by given the period
func (p *AmigaPeriod) GetSamplerAdd(samplerSpeed float64) float64 {
	period := float64(*p)
	if period == 0 {
		return 0
	}
	return samplerSpeed / period
}

var semitonePeriodTable = [...]float32{27392, 25856, 24384, 23040, 21696, 20480, 19328, 18240, 17216, 16256, 15360, 14496}

// CalcSemitonePeriod calculates the semitone period for S3M notes
func CalcSemitonePeriod(semi note.Semitone, c2spd note.C2SPD) note.Period {
	key := int(semi.Key())
	octave := int(semi.Octave())

	if key >= len(semitonePeriodTable) {
		return nil
	}

	if c2spd == 0 {
		c2spd = note.C2SPD(s3mfile.DefaultC2Spd)
	}

	period := (AmigaPeriod(floatDefaultC2Spd*semitonePeriodTable[key]) / AmigaPeriod(uint32(c2spd)<<octave))
	period = period.AddInteger(0)
	return &period
}

// CalcFinetuneC2Spd calculates a new C2SPD after a finetune adjustment
func CalcFinetuneC2Spd(c2spd note.C2SPD, finetune int8) note.C2SPD {
	if finetune == 0 {
		return c2spd
	}

	o := 5
	st := note.Semitone(o * 12) // C-5
	stShift := int8(finetune / 16)
	if stShift >= 0 {
		st += note.Semitone(stShift)
	} else {
		st -= note.Semitone(-stShift)
	}
	period0 := CalcSemitonePeriod(st, c2spd)
	period1 := CalcSemitonePeriod(st+1, c2spd)
	fFt := float64(finetune) / 16
	iFt := math.Trunc(fFt)
	f := fFt - iFt
	period := period0.Lerp(f, period1)
	return note.C2SPD(FrequencyFromPeriod(period))
}

// VolumeFromS3M converts an S3M volume to a player volume
func VolumeFromS3M(vol s3mfile.Volume) volume.Volume {
	var v volume.Volume
	switch {
	case vol == s3mfile.EmptyVolume:
		v = volume.VolumeUseInstVol
	case vol >= 63:
		v = volume.Volume(63.0) / 64.0
	case vol < 63:
		v = volume.Volume(vol) / 64.0
	default:
		v = 0.0
	}
	return v
}

// VolumeToS3M converts a player volume to an S3M volume
func VolumeToS3M(v volume.Volume) s3mfile.Volume {
	switch {
	case v == volume.VolumeUseInstVol:
		return s3mfile.EmptyVolume
	default:
		return s3mfile.Volume(v * 64.0)
	}
}

// VolumeFromS3M8BitSample converts an S3M 8-bit sample volume to a player volume
func VolumeFromS3M8BitSample(vol uint8) volume.Volume {
	return (volume.Volume(vol) - 128.0) / 128.0
}

// VolumeFromS3M16BitSample converts an S3M 16-bit sample volume to a player volume
func VolumeFromS3M16BitSample(vol uint16) volume.Volume {
	return (volume.Volume(vol) - 32768.0) / 32768.0
}

// PanningFromS3M returns a radian panning position from an S3M panning value
func PanningFromS3M(pos uint8) panning.Position {
	return panning.MakeStereoPosition(float32(pos), 0, 0x0F)
}

// NoteFromS3MNote converts an S3M file note into a player note
func NoteFromS3MNote(sn s3mfile.Note) note.Note {
	switch {
	case sn == s3mfile.EmptyNote:
		return note.EmptyNote
	case sn == s3mfile.StopNote:
		return note.StopNote
	default:
		k := uint8(sn.Key()) & 0x0f
		o := uint8(sn.Octave()) & 0x0f
		if k < 12 && o < 10 {
			s := note.Semitone(o*12 + k)
			return note.NewNote(s)
		}
	}
	return note.InvalidNote
}

// FrequencyFromSemitone returns the frequency from the semitone (and c2spd)
func FrequencyFromSemitone(semitone note.Semitone, c2spd note.C2SPD) float32 {
	period := CalcSemitonePeriod(semitone, c2spd)
	return FrequencyFromPeriod(period)
}

// FrequencyFromPeriod returns the frequency from the period
func FrequencyFromPeriod(period note.Period) float32 {
	if p, ok := period.(*AmigaPeriod); ok {
		return S3MBaseClock / float32(*p)
	}
	return 0
}
