package util

import (
	"math"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/note"
)

const (
	// DefaultC2Spd is the default C2SPD for XM samples
	DefaultC2Spd = 8363

	floatDefaultC2Spd = float32(DefaultC2Spd)
	c2Period          = float32(1712)

	// XMBaseClock is the base clock speed of xm files
	XMBaseClock = floatDefaultC2Spd * c2Period
)

var (
	// DefaultVolume is the default volume value for most everything in xm format
	DefaultVolume = VolumeFromXm(0x10 + 0x40)

	// DefaultMixingVolume is the default mixing volume
	DefaultMixingVolume = volume.Volume(0x30) / 0x80

	// DefaultPanningLeft is the default panning value for left channels
	DefaultPanningLeft = PanningFromXm(0x30)
	// DefaultPanning is the default panning value for unconfigured channels
	DefaultPanning = PanningFromXm(0x80)
	// DefaultPanningRight is the default panning value for right channels
	DefaultPanningRight = PanningFromXm(0xC0)
)

var semitonePeriodTable = [...]float32{27392, 25856, 24384, 23040, 21696, 20480, 19328, 18240, 17216, 16256, 15360, 14496}

// CalcSemitonePeriod calculates the semitone period for xm notes
func CalcSemitonePeriod(semi note.Semitone, ft note.Finetune, c2spd note.C2SPD, linearFreqSlides bool) note.Period {
	if linearFreqSlides {
		return &LinearPeriod{
			Semitone: semi,
			Finetune: ft,
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

// VolumeFromXm converts an xm volume to a player volume
func VolumeFromXm(vol uint8) volume.Volume {
	var v volume.Volume
	switch {
	case vol >= 0x10 && vol <= 0x50:
		v = volume.Volume(vol-0x10) / 64.0
	default:
		v = volume.VolumeUseInstVol
	}
	return v
}

// VolumeToXm converts a player volume to an xm volume
func VolumeToXm(v volume.Volume) uint8 {
	switch {
	case v == volume.VolumeUseInstVol:
		return 0
	default:
		return uint8(v*64.0) + 0x10
	}
}

// VolumeFromXm8BitSample converts an xm 8-bit sample volume to a player volume
func VolumeFromXm8BitSample(vol uint8) volume.Volume {
	return volume.Volume(int8(vol)) / 128.0
}

// VolumeFromXm16BitSample converts an xm 16-bit sample volume to a player volume
func VolumeFromXm16BitSample(vol uint16) volume.Volume {
	return volume.Volume(int16(vol)) / 32768.0
}

// PanningFromXm returns a radian panning position from an xm panning value
func PanningFromXm(pos uint8) panning.Position {
	return panning.MakeStereoPosition(float32(pos), 0, 0xFF)
}

// PanningToXm returns the xm panning value for a radian panning position
func PanningToXm(pan panning.Position) uint8 {
	return uint8(panning.FromStereoPosition(pan, 0, 0xFF))
}

// NoteFromXmNote converts an xm file note into a player note
func NoteFromXmNote(xn uint8) note.Note {
	switch {
	case xn == 97:
		return note.StopNote
	case xn == 0:
		return note.EmptyNote
	case xn > 97: // invalid
		return note.InvalidNote
	}

	an := uint8(xn - 1)
	s := note.Semitone(an)
	return note.NewNote(s)
}

// FrequencyFromSemitone returns the frequency from the semitone (and c2spd)
func FrequencyFromSemitone(semitone note.Semitone, c2spd note.C2SPD, linearFreqSlides bool) float32 {
	period := CalcSemitonePeriod(semitone, 0, c2spd, linearFreqSlides)
	return float32(period.GetFrequency())
}

// CalcLinearPeriod calculates a period for a linear frequency slide
func CalcLinearPeriod(n note.Semitone, ft note.Finetune, c2spd note.C2SPD) note.Period {
	nsf := int(n)*64 + int(ft)

	linFreq := math.Pow(2, float64(nsf)/768)

	period := AmigaPeriod(float64(semitonePeriodTable[0]) / linFreq)
	return &period
}

// ToLinearPeriod returns the linear frequency period for a given period
func ToLinearPeriod(p note.Period) note.Period {
	switch pp := p.(type) {
	case *LinearPeriod:
		return pp
	case *AmigaPeriod:
		c5 := note.NewSemitone(note.KeyC, 5)
		freq := pp.GetFrequency()
		for f := -3840; f < 3840; f++ {
			lp := LinearPeriod{
				Semitone: c5,
				Finetune: note.Finetune(f),
			}
			if lp.GetFrequency() >= freq {
				return &lp
			}
		}
	}
	return nil
}
