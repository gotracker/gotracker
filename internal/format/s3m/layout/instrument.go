package layout

import (
	"math"
	"time"

	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// InstrumentDataIntf is the interface to implementation-specific functions on an instrument
type InstrumentDataIntf interface {
	GetSample(*InstrumentOnChannel, sampling.Pos) volume.Matrix

	Initialize(*InstrumentOnChannel) error
	SetKeyOn(*InstrumentOnChannel, note.Period, bool)
	GetKeyOn(*InstrumentOnChannel) bool
	Update(*InstrumentOnChannel, time.Duration)
}

// InstrumentOnChannel is an instance of the instrument on a particular output channel
type InstrumentOnChannel struct {
	intf.InstrumentOnChannel

	Instrument       *Instrument
	OutputChannelNum int
	Volume           volume.Volume
	Data             interface{}
	Filter           intf.Filter
	Playback         intf.Playback
	Period           note.Period
}

// Instrument is the mildly-decoded S3M instrument/sample header
type Instrument struct {
	intf.Instrument

	Filename string
	Name     string
	Inst     InstrumentDataIntf
	ID       channel.S3MInstrumentID
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
	case *InstrumentOPL2:
		return sampling.Pos{Pos: math.MaxInt64, Frac: 0}
	case *InstrumentPCM:
		return sampling.Pos{Pos: si.Length}
	default:
	}
	return sampling.Pos{}
}

// InstantiateOnChannel takes an instrument and loads it onto an output channel
func (inst *Instrument) InstantiateOnChannel(channelIdx int, filter intf.Filter) intf.InstrumentOnChannel {
	ioc := InstrumentOnChannel{
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

// GetSample returns the sample at position `pos` in the instrument
func (inst *InstrumentOnChannel) GetSample(pos sampling.Pos) volume.Matrix {
	if inst.Instrument != nil && inst.Instrument.Inst != nil {
		dry := inst.Instrument.Inst.GetSample(inst, pos)
		if inst.Filter == nil {
			return dry
		}
		wet := inst.Filter.Filter(dry)
		return wet
	}
	return nil
}

// GetInstrument returns the instrument that's on this instance
func (inst *InstrumentOnChannel) GetInstrument() intf.Instrument {
	return inst.Instrument
}

// SetKeyOn sets the key on flag for the instrument
func (inst *InstrumentOnChannel) SetKeyOn(period note.Period, on bool) {
	if inst.Instrument != nil && inst.Instrument.Inst != nil {
		inst.Instrument.Inst.SetKeyOn(inst, period, on)
	}
}

// GetKeyOn gets the key on flag for the instrument
func (inst *InstrumentOnChannel) GetKeyOn() bool {
	if inst.Instrument != nil && inst.Instrument.Inst != nil {
		return inst.Instrument.Inst.GetKeyOn(inst)
	}
	return false
}

// Update advances time by the amount specified by `tickDuration`
func (inst *InstrumentOnChannel) Update(tickDuration time.Duration) {
	if inst.Instrument != nil && inst.Instrument.Inst != nil {
		inst.Instrument.Inst.Update(inst, tickDuration)
	}
}

// SetFilter sets the active filter on the instrument (which should be the same as what's on the channel)
func (inst *InstrumentOnChannel) SetFilter(filter intf.Filter) {
	inst.Filter = filter
}

// SetVolume sets the active instrument on channel's volume
func (inst *InstrumentOnChannel) SetVolume(vol volume.Volume) {
	inst.Volume = vol
}

// SetPeriod sets the active instrument on channel's period
func (inst *InstrumentOnChannel) SetPeriod(period note.Period) {
	inst.Period = period
}
