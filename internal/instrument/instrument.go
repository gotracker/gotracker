package instrument

import (
	"math"
	"time"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/oscillator"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
	"gotracker/internal/player/state"
)

// DataIntf is the interface to implementation-specific functions on an instrument
type DataIntf interface {
	GetSample(intf.NoteControl, sampling.Pos) volume.Matrix
	GetCurrentPanning(intf.NoteControl) panning.Position
	SetEnvelopePosition(intf.NoteControl, int)
	Initialize(intf.NoteControl) error
	Attack(intf.NoteControl)
	Release(intf.NoteControl)
	GetKeyOn(intf.NoteControl) bool
	Update(intf.NoteControl, time.Duration)
}

// AutoVibrato is the setting and memory for the auto-vibrato system
type AutoVibrato struct {
	Enabled           bool
	Sweep             uint8
	WaveformSelection uint8
	Depth             uint8
	Rate              uint8
}

// Instrument is the mildly-decoded instrument/sample header
type Instrument struct {
	Filename           string
	Name               string
	Inst               DataIntf
	ID                 intf.InstrumentID
	C2Spd              note.C2SPD
	Volume             volume.Volume
	RelativeNoteNumber int8
	Finetune           note.Finetune
	AutoVibrato        AutoVibrato
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
	case *PCM:
		return si.LoopMode != LoopModeDisabled
	default:
	}
	return false
}

// GetLoopBegin returns the loop start position
func (inst *Instrument) GetLoopBegin() sampling.Pos {
	switch si := inst.Inst.(type) {
	case *PCM:
		return sampling.Pos{Pos: si.LoopBegin}
	default:
	}
	return sampling.Pos{}
}

// GetLoopEnd returns the loop end position
func (inst *Instrument) GetLoopEnd() sampling.Pos {
	switch si := inst.Inst.(type) {
	case *PCM:
		return sampling.Pos{Pos: si.LoopEnd}
	default:
	}
	return sampling.Pos{}
}

// GetLength returns the length of the instrument
func (inst *Instrument) GetLength() sampling.Pos {
	switch si := inst.Inst.(type) {
	case *OPL2:
		return sampling.Pos{Pos: math.MaxInt64, Frac: 0}
	case *PCM:
		return sampling.Pos{Pos: si.Length}
	default:
	}
	return sampling.Pos{}
}

// SetFinetune sets the finetune value on the instrument
func (inst *Instrument) SetFinetune(ft note.Finetune) {
	inst.Finetune = ft
}

// GetFinetune returns the finetune value on the instrument
func (inst *Instrument) GetFinetune() note.Finetune {
	return inst.Finetune
}

// InstantiateOnChannel takes an instrument and loads it onto an output channel
func (inst *Instrument) InstantiateOnChannel(oc *intf.OutputChannel) intf.NoteControl {
	ioc := state.NoteControl{
		Output: oc,
	}
	ioc.Instrument = inst

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
		if inst.AutoVibrato.Enabled {
			if ncav := nc.GetAutoVibratoState(); ncav != nil {
				ncav.Osc.Table = oscillator.WaveTableSelect(inst.AutoVibrato.WaveformSelection)
				ncav.Osc.Advance(int(inst.AutoVibrato.Rate))
				ncav.Ticks++
				d := float32(inst.AutoVibrato.Depth) / 64
				if inst.AutoVibrato.Sweep > 0 && ncav.Ticks < int(inst.AutoVibrato.Sweep) {
					d *= float32(ncav.Ticks) / float32(inst.AutoVibrato.Sweep)
				}
				if ncs := nc.GetPlaybackState(); ncs != nil {
					pd := note.PeriodDelta(ncav.Osc.GetWave(d))
					ncs.Period = ncs.Period.Add(pd)
				}
			}
		}
		ii.Update(nc, tickDuration)
	}
}

// SetEnvelopePosition sets the envelope position for the instrument
func (inst *Instrument) SetEnvelopePosition(nc intf.NoteControl, ticks int) {
	if ii := inst.Inst; ii != nil {
		ii.SetEnvelopePosition(nc, ticks)
	}
}
