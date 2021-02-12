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
	if semi == note.UnchangedSemitone {
		panic("how?")
	}
	if linearFreqSlides {
		nft := int(semi)*64 + int(ft)
		return &LinearPeriod{
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

// VolumeXM is a helpful converter from the XM range of 0-64 into a volume
type VolumeXM uint8

const cVolumeXMCoeff = volume.Volume(1) / 0x40

// Volume returns the volume from the internal format
func (v VolumeXM) Volume() volume.Volume {
	return volume.Volume(v) * cVolumeXMCoeff
}

// ToVolumeXM returns the VolumeXM representation of a volume
func ToVolumeXM(v volume.Volume) VolumeXM {
	return VolumeXM(v * 0x40)
}

// VolEffect holds the data related to volume and effects from the volume data channel
type VolEffect uint8

// IsVolume returns true if the VolEffect describes a volume value
func (v VolEffect) IsVolume() bool {
	return v == 0x00 || v >= 0x10 && v <= 0x50
}

// Volume returns the value from the volume portion of the range
func (v VolEffect) Volume() volume.Volume {
	if v == 0x00 {
		return volume.VolumeUseInstVol
	}
	return VolumeXM(v - 0x10).Volume()
}

// VolumeFromXm converts an xm volume to a player volume
func VolumeFromXm(vol VolEffect) volume.Volume {
	if vol.IsVolume() {
		return vol.Volume()
	}
	panic("unexpected conversion of non-volume value")
}

// VolumeToXm converts a player volume to an xm volume
func VolumeToXm(v volume.Volume) VolEffect {
	switch {
	case v == volume.VolumeUseInstVol:
		return 0
	case v >= 0 && v <= 1:
		return VolEffect(v*0x40) + 0x10
	default:
		panic("volume out of range for conversion")
	}
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
		return note.ReleaseNote
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

// ToAmigaPeriod calculates an amiga period for a linear finetune period
func ToAmigaPeriod(finetunes note.Finetune, c2spd note.C2SPD) AmigaPeriod {
	linFreq := float64(c2spd) * math.Pow(2, float64(finetunes)/768) / DefaultC2Spd

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
