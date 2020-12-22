package layout

import (
	"github.com/heucuva/gomixing/sampling"
	"github.com/heucuva/gomixing/volume"

	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// InstrumentDataIntf is the interface to implementation-specific functions on an instrument
type InstrumentDataIntf interface {
	GetSample(*InstrumentOnChannel, sampling.Pos) volume.Matrix

	Initialize(*InstrumentOnChannel) error
	SetKeyOn(*InstrumentOnChannel, note.Semitone, bool)
	GetKeyOn(*InstrumentOnChannel) bool
}

// InstrumentOnChannel is an instance of the instrument on a particular output channel
type InstrumentOnChannel struct {
	intf.InstrumentOnChannel

	Instrument       *Instrument
	OutputChannelNum int
	Data             interface{}
}

// Instrument is the mildly-decoded S3M instrument/sample header
type Instrument struct {
	intf.Instrument

	Filename string
	Name     string
	Inst     InstrumentDataIntf
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

// InstantiateOnChannel takes an instrument and loads it onto an output channel
func (inst *Instrument) InstantiateOnChannel(channelIdx int) intf.InstrumentOnChannel {
	ioc := InstrumentOnChannel{
		OutputChannelNum: channelIdx,
		Instrument:       inst,
	}

	if inst.Inst != nil {
		inst.Inst.Initialize(&ioc)
	}

	return &ioc
}

// GetID returns the instrument number (1-based)
func (inst *Instrument) GetID() int {
	return int(inst.ID)
}

// GetSample returns the sample at position `pos` in the instrument
func (inst *InstrumentOnChannel) GetSample(pos sampling.Pos) volume.Matrix {
	if inst.Instrument != nil && inst.Instrument.Inst != nil {
		return inst.Instrument.Inst.GetSample(inst, pos)
	}
	return nil
}

// GetInstrument returns the instrument that's on this instance
func (inst *InstrumentOnChannel) GetInstrument() intf.Instrument {
	return inst.Instrument
}

// SetKeyOn sets the key on flag for the instrument
func (inst *InstrumentOnChannel) SetKeyOn(semitone note.Semitone, on bool) {

	if inst.Instrument != nil && inst.Instrument.Inst != nil {
		inst.Instrument.Inst.SetKeyOn(inst, semitone, on)
	}
}

// GetKeyOn gets the key on flag for the instrument
func (inst *InstrumentOnChannel) GetKeyOn() bool {
	if inst.Instrument != nil && inst.Instrument.Inst != nil {
		return inst.Instrument.Inst.GetKeyOn(inst)
	}
	return false
}
