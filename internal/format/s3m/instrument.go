package s3m

import (
	"gotracker/internal/format/s3m/util"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/volume"
)

// Instrument is the mildly-decoded S3M instrument/sample header
type Instrument struct {
	intf.Instrument
	//Info      SCRSHeader
	Filename    string
	Name        string
	Sample      []uint8
	Length      int
	ID          uint8
	C2Spd       note.C2SPD
	Volume      volume.Volume
	Looped      bool
	LoopBegin   float32
	LoopEnd     float32
	NumChannels int
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
func (inst *Instrument) GetLoopBegin() float32 {
	return inst.LoopBegin
}

// GetLoopEnd returns the loop end position
func (inst *Instrument) GetLoopEnd() float32 {
	return inst.LoopEnd
}

// GetLength returns the length of the instrument
func (inst *Instrument) GetLength() float32 {
	return float32(inst.Length)
}

// GetSample returns the sample at position `pos` in the instrument
func (inst *Instrument) GetSample(pos float32) volume.VolumeMatrix {
	o := make(volume.VolumeMatrix, inst.NumChannels)
	if pos < 0 {
		return o
	}

	if inst.Looped {
		pos = util.CalcLoopedSamplePos(pos, inst.LoopBegin, inst.LoopEnd)
	}

	if int(pos) >= inst.Length {
		return o
	}

	i := int(pos)
	if i >= 0 {
		t := pos - float32(i)
		for c := 0; c < inst.NumChannels; c++ {
			v0 := util.VolumeFromS3M8BitSample(inst.Sample[i*inst.NumChannels+c])
			if t == 0 || i == inst.Length-1 {
				o[c] = v0
			} else {
				v1 := util.VolumeFromS3M8BitSample(inst.Sample[(i+1)*inst.NumChannels+c])
				o[c] = v0 + volume.Volume(t)*(v1-v0)
			}
		}
	}
	return o
}

// GetID returns the instrument number (1-based)
func (inst *Instrument) GetID() int {
	return int(inst.ID)
}
