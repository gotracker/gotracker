package layout

import (
	"math"
	"time"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/instrument"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/state"
)

// Instrument is the mildly-decoded S3M instrument/sample header
type Instrument struct {
	intf.Instrument

	Filename string
	Name     string
	Inst     instrument.DataIntf
	ID       channel.S3MInstrumentID
	C2Spd    note.C2SPD
	Volume   volume.Volume
	Finetune note.Finetune
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

// GetDefaultVolume returns the default volume value for the instrument
func (inst *Instrument) GetDefaultVolume() volume.Volume {
	return inst.Volume
}

// SetFinetune sets the finetune value on the instrument
func (inst *Instrument) SetFinetune(ft note.Finetune) {
	inst.Finetune = ft
}

// GetFinetune returns the finetune value on the instrument
func (inst *Instrument) GetFinetune() note.Finetune {
	return inst.Finetune
}

// IsLooped returns true if the instrument has the loop flag set
func (inst *Instrument) IsLooped() bool {
	switch si := inst.Inst.(type) {
	case *instrument.PCM:
		return si.LoopMode != instrument.LoopModeDisabled
	default:
	}
	return false
}

// GetLoopBegin returns the loop start position
func (inst *Instrument) GetLoopBegin() sampling.Pos {
	switch si := inst.Inst.(type) {
	case *instrument.PCM:
		return sampling.Pos{Pos: si.LoopBegin}
	default:
	}
	return sampling.Pos{}
}

// GetLoopEnd returns the loop end position
func (inst *Instrument) GetLoopEnd() sampling.Pos {
	switch si := inst.Inst.(type) {
	case *instrument.PCM:
		return sampling.Pos{Pos: si.LoopEnd}
	default:
	}
	return sampling.Pos{}
}

// GetLength returns the length of the instrument
func (inst *Instrument) GetLength() sampling.Pos {
	switch si := inst.Inst.(type) {
	case *instrument.OPL2:
		return sampling.Pos{Pos: math.MaxInt64, Frac: 0}
	case *instrument.PCM:
		return sampling.Pos{Pos: si.Length}
	default:
	}
	return sampling.Pos{}
}

// InstantiateOnChannel takes an instrument and loads it onto an output channel
func (inst *Instrument) InstantiateOnChannel(channelIdx int, filter intf.Filter) intf.NoteControl {
	ioc := state.NoteControl{
		OutputChannelNum: channelIdx,
		Instrument:       inst,
		Filter:           filter,
	}

	if inst.Inst != nil {
		inst.Inst.Initialize(&ioc)
	}

	return &ioc
}

// GetID returns the instrument number (1-based)
func (inst *Instrument) GetID() intf.InstrumentID {
	return inst.ID
}

// GetSemitoneShift returns the amount of semitones worth of shift to play the instrument at
func (inst *Instrument) GetSemitoneShift() int8 {
	return 0
}

// GetSample returns a sample from the instrument at the specified position
func (inst *Instrument) GetSample(nc intf.NoteControl, pos sampling.Pos) volume.Matrix {
	if ii := inst.Inst; ii != nil {
		return ii.GetSample(nc, pos)
	}
	return nil
}

// GetCurrentPanning returns the panning envelope position
func (inst *Instrument) GetCurrentPanning(nc intf.NoteControl) panning.Position {
	if ii := inst.Inst; ii != nil {
		return ii.GetCurrentPanning(nc)
	}
	return panning.CenterAhead
}

// Attack sets the key-on flag for the instrument
func (inst *Instrument) Attack(nc intf.NoteControl) {
	if ii := inst.Inst; ii != nil {
		ii.Attack(nc)
	}
}

// Release clears the key-on flag for the instrument
func (inst *Instrument) Release(nc intf.NoteControl) {
	if ii := inst.Inst; ii != nil {
		ii.Release(nc)
	}
}

// NoteCut clears the key-on flag for the instrument and stops any output from it
func (inst *Instrument) NoteCut(nc intf.NoteControl) {
	if ii := inst.Inst; ii != nil {
		ii.NoteCut(nc)
	}
}

// GetKeyOn returns the key-on flag state for the instrument
func (inst *Instrument) GetKeyOn(nc intf.NoteControl) bool {
	if ii := inst.Inst; ii != nil {
		return ii.GetKeyOn(nc)
	}
	return false
}

// Update updates the instrument
func (inst *Instrument) Update(nc intf.NoteControl, tickDuration time.Duration) {
	if ii := inst.Inst; ii != nil {
		ii.Update(nc, tickDuration)
	}
}
