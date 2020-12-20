package layout

import (
	"github.com/heucuva/gomixing/sampling"
	"github.com/heucuva/gomixing/volume"

	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// Instrument is the mildly-decoded S3M instrument/sample header
type Instrument struct {
	intf.Instrument

	Filename string
	Name     string
	Inst     intf.Instrument
	ID       uint8
	C2Spd    note.C2SPD
	Volume   volume.Volume
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
	switch si := inst.Inst.(type) {
	case *InstrumentPCM:
		return si.Looped
	default:
	}
	return false
}

// GetLoopBegin returns the loop start position
func (inst *Instrument) GetLoopBegin() sampling.Pos {
	switch si := inst.Inst.(type) {
	case *InstrumentPCM:
		return sampling.Pos{Pos: si.LoopBegin}
	default:
	}
	return sampling.Pos{}
}

// GetLoopEnd returns the loop end position
func (inst *Instrument) GetLoopEnd() sampling.Pos {
	switch si := inst.Inst.(type) {
	case *InstrumentPCM:
		return sampling.Pos{Pos: si.LoopEnd}
	default:
	}
	return sampling.Pos{}
}

// GetLength returns the length of the instrument
func (inst *Instrument) GetLength() sampling.Pos {
	switch si := inst.Inst.(type) {
	case *InstrumentPCM:
		return sampling.Pos{Pos: si.Length}
	default:
	}
	return sampling.Pos{}
}

// GetSample returns the sample at position `pos` in the instrument
func (inst *Instrument) GetSample(pos sampling.Pos) volume.Matrix {
	if inst.Inst != nil {
		return inst.Inst.GetSample(pos)
	}
	return nil
}

// GetID returns the instrument number (1-based)
func (inst *Instrument) GetID() int {
	return int(inst.ID)
}
