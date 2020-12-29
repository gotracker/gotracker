package util

import (
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
	DefaultVolume = VolumeFromXm(0x50)

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
func CalcSemitonePeriod(semi note.Semitone, c2spd note.C2SPD) note.Period {
	key := int(semi) % len(semitonePeriodTable)
	octave := uint(int(semi) / len(semitonePeriodTable))

	if key >= len(semitonePeriodTable) {
		return 0
	}

	if c2spd == 0 {
		c2spd = note.C2SPD(DefaultC2Spd)
	}

	period := (note.Period(floatDefaultC2Spd*semitonePeriodTable[key]) / note.Period(uint32(c2spd)<<octave))
	return period.AddInteger(0)
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

// NoteFromXmNote converts an xm file note into a player note
func NoteFromXmNote(xn uint8) note.Note {
	switch {
	case xn == 97:
		return note.StopNote
	case xn == 0:
		return note.EmptyNote
	case xn > 97: // invalid
		return note.Note(note.KeyInvalid1)
	}

	an := uint8(xn - 1)
	k := an % 12
	o := an / 12
	return note.Note(o<<4 | k)
}

// FrequencyFromSemitone returns the frequency from the semitone (and c2spd)
func FrequencyFromSemitone(semitone note.Semitone, c2spd note.C2SPD) float32 {
	period := CalcSemitonePeriod(semitone, c2spd)
	return FrequencyFromPeriod(period)
}

// FrequencyFromPeriod returns the frequency from the period
func FrequencyFromPeriod(period note.Period) float32 {
	return XMBaseClock / float32(period)
}
