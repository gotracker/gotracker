package layout

import (
	"math"
	"time"

	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/state"
)

// InstrumentDataIntf is the interface to implementation-specific functions on an instrument
type InstrumentDataIntf interface {
	GetSample(intf.NoteControl, sampling.Pos) volume.Matrix

	Initialize(intf.NoteControl) error
	Attack(intf.NoteControl)
	Release(intf.NoteControl)
	NoteCut(intf.NoteControl)
	GetKeyOn(intf.NoteControl) bool
	Update(intf.NoteControl, time.Duration)
}

// Instrument is the mildly-decoded XM instrument/sample header
type Instrument struct {
	intf.Instrument

	Filename           string
	Name               string
	Inst               InstrumentDataIntf
	ID                 channel.SampleID
	C2Spd              note.C2SPD
	Volume             volume.Volume
	RelativeNoteNumber int8
	Finetune           int8
}

// IsInvalid always returns false (valid)
func (inst *Instrument) IsInvalid() bool {
	return false
}

// GetC2Spd returns the C2SPD value for the instrument
// This may get mutated if a finetune effect is processed
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
	case *InstrumentOPL2:
		return sampling.Pos{Pos: math.MaxInt64, Frac: 0}
	case *InstrumentPCM:
		return sampling.Pos{Pos: si.Length}
	default:
	}
	return sampling.Pos{}
}

// SetFinetune sets the finetune value on the instrument
func (inst *Instrument) SetFinetune(ft int8) {
	inst.Finetune = ft
}

// GetFinetune returns the finetune value on the instrument
func (inst *Instrument) GetFinetune() int8 {
	return inst.Finetune
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
	return inst.RelativeNoteNumber
}

// GetSample returns a sample from the instrument at the specified position
func (inst *Instrument) GetSample(nc intf.NoteControl, pos sampling.Pos) volume.Matrix {
	if ii := inst.Inst; ii != nil {
		return ii.GetSample(nc, pos)
	}
	return nil
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
