package instrument

import (
	"math"

	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/optional"
	"gotracker/internal/player/intf"
	voiceIntf "gotracker/internal/player/intf/voice"
	"gotracker/internal/player/note"
)

// StaticValues are the static values associated with an instrument
type StaticValues struct {
	Filename           string
	Name               string
	ID                 intf.InstrumentID
	Volume             volume.Volume
	RelativeNoteNumber int8
	AutoVibrato        voiceIntf.AutoVibrato
	NewNoteAction      note.Action
	Finetune           note.Finetune
	FilterFactory      intf.FilterFactory
}

// Instrument is the mildly-decoded instrument/sample header
type Instrument struct {
	Static   StaticValues
	Inst     intf.InstrumentDataIntf
	C2Spd    note.C2SPD
	Finetune optional.Value //note.Finetune
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
	return inst.Static.Volume
}

// GetLength returns the length of the instrument
func (inst *Instrument) GetLength() sampling.Pos {
	switch si := inst.Inst.(type) {
	case *OPL2:
		return sampling.Pos{Pos: math.MaxInt64, Frac: 0}
	case *PCM:
		return sampling.Pos{Pos: si.Sample.Length()}
	default:
	}
	return sampling.Pos{}
}

// SetFinetune sets the finetune value on the instrument
func (inst *Instrument) SetFinetune(ft note.Finetune) {
	inst.Finetune.Set(ft)
}

// GetFinetune returns the finetune value on the instrument
func (inst *Instrument) GetFinetune() note.Finetune {
	if ft, ok := inst.Finetune.GetFinetune(); ok {
		return ft
	}
	return inst.Static.Finetune
}

// GetID returns the instrument number (1-based)
func (inst *Instrument) GetID() intf.InstrumentID {
	return inst.Static.ID
}

// GetSemitoneShift returns the amount of semitones worth of shift to play the instrument at
func (inst *Instrument) GetSemitoneShift() int8 {
	return inst.Static.RelativeNoteNumber
}

// GetKind returns the kind of the instrument
func (inst *Instrument) GetKind() intf.InstrumentKind {
	switch inst.Inst.(type) {
	case *PCM:
		return intf.InstrumentKindPCM
	case *OPL2:
		return intf.InstrumentKindOPL2
	}
	return intf.InstrumentKindPCM
}

// GetNewNoteAction returns the NewNoteAction associated to the instrument
func (inst *Instrument) GetNewNoteAction() note.Action {
	return inst.Static.NewNoteAction
}

// GetData returns the instrument-specific data interface
func (inst *Instrument) GetData() intf.InstrumentDataIntf {
	return inst.Inst
}

// GetChannelFilterFactory returns the factory for the channel filter
func (inst *Instrument) GetChannelFilterFactory() intf.FilterFactory {
	return inst.Static.FilterFactory
}

// GetAutoVibrato returns the settings for the autovibrato system
func (inst *Instrument) GetAutoVibrato() voiceIntf.AutoVibrato {
	return inst.Static.AutoVibrato
}

// IsReleaseNote returns true if the note is a release (Note-Off)
func (inst *Instrument) IsReleaseNote(n note.Note) bool {
	switch n.Type() {
	case note.SpecialTypeStopOrRelease:
		return inst.GetKind() == intf.InstrumentKindOPL2
	}
	return n.IsRelease()
}

// IsStopNote returns true if the note is a stop (Note-Cut)
func (inst *Instrument) IsStopNote(n note.Note) bool {
	switch n.Type() {
	case note.SpecialTypeStopOrRelease:
		return inst.GetKind() == intf.InstrumentKindPCM
	}
	return n.IsRelease()
}
