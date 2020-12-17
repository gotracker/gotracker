package s3m

import (
	"encoding/binary"

	"github.com/heucuva/gomixing/sampling"
	"github.com/heucuva/gomixing/volume"

	"gotracker/internal/format/s3m/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// Instrument is the mildly-decoded S3M instrument/sample header
type Instrument struct {
	intf.Instrument
	//Info      SCRSHeader
	Filename      string
	Name          string
	Sample        []uint8
	Length        int
	ID            uint8
	C2Spd         note.C2SPD
	Volume        volume.Volume
	Looped        bool
	LoopBegin     int
	LoopEnd       int
	NumChannels   int
	BitsPerSample int
}

// IsInvalid always returns false (valid)
func (inst *Instrument) IsInvalid() bool {
	return false
}

// GetC2Spd returns the C2SPD value for the instrument
// This may get mutated if a finetune command is processed
func (inst *Instrument) GetC2Spd() note.C2SPD {
	return inst.C2Spd
}

// SetC2Spd sets the C2SPD value for the instrument
func (inst *Instrument) SetC2Spd(c2spd note.C2SPD) {
	inst.C2Spd = c2spd
}

// GetVolume returns the default volume value for the instrument
func (inst *Instrument) GetVolume() volume.Volume {
	return inst.Volume
}

// IsLooped returns true if the instrument has the loop flag set
func (inst *Instrument) IsLooped() bool {
	return inst.Looped
}

// GetLoopBegin returns the loop start position
func (inst *Instrument) GetLoopBegin() sampling.Pos {
	return sampling.Pos{Pos: inst.LoopBegin}
}

// GetLoopEnd returns the loop end position
func (inst *Instrument) GetLoopEnd() sampling.Pos {
	return sampling.Pos{Pos: inst.LoopEnd}
}

// GetLength returns the length of the instrument
func (inst *Instrument) GetLength() sampling.Pos {
	return sampling.Pos{Pos: inst.Length}
}

// GetSample returns the sample at position `pos` in the instrument
func (inst *Instrument) GetSample(pos sampling.Pos) volume.Matrix {
	v0 := inst.getConvertedSample(pos.Pos)
	if pos.Frac == 0 {
		return v0
	}
	v1 := inst.getConvertedSample(pos.Pos + 1)
	for c, s := range v1 {
		v0[c] += volume.Volume(pos.Frac) * (s - v0[c])
	}
	return v0
}

func (inst *Instrument) getConvertedSample(pos int) volume.Matrix {
	if inst.Looped {
		pos = inst.calcLoopedSamplePos(pos)
	}
	if pos < 0 || pos >= inst.Length {
		return volume.Matrix{}
	}
	o := make(volume.Matrix, inst.NumChannels)
	for c := 0; c < inst.NumChannels; c++ {
		switch inst.BitsPerSample {
		case 8:
			o[c] = util.VolumeFromS3M8BitSample(inst.Sample[pos+c])
		case 16:
			s := binary.LittleEndian.Uint16(inst.Sample[pos+c:])
			o[c] = util.VolumeFromS3M16BitSample(s)
		}
	}
	return o
}

func (inst *Instrument) calcLoopedSamplePos(pos int) int {
	// |start--------------------------------------------------end|   <= on playthrough 1, whole sample plays
	// |----------------|loopBegin---------loopEnd|---------------|   <= on playthrough 2+, only the part that loops plays
	if pos < 0 {
		return 0
	}
	if pos < inst.Length {
		return pos
	}
	loopedPos := pos
	for {
		lastLoopedPos := loopedPos
		loopedPos += inst.LoopBegin - inst.LoopEnd
		if loopedPos < inst.LoopEnd {
			return loopedPos
		}
		if lastLoopedPos == loopedPos {
			return 0 // do not allow infinite loop
		}
	}
}

// GetID returns the instrument number (1-based)
func (inst *Instrument) GetID() int {
	return int(inst.ID)
}
