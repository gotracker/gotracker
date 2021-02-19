package intf

import (
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	voiceIntf "gotracker/internal/player/intf/voice"
	"gotracker/internal/player/note"
)

// InstrumentID is an identifier for an instrument/sample that means something to the format
type InstrumentID interface {
	IsEmpty() bool
}

// InstrumentDataIntf is the interface to implementation-specific functions on an instrument
type InstrumentDataIntf interface{}

// InstrumentKind defines the kind of instrument
type InstrumentKind int

const (
	// InstrumentKindPCM defines a PCM instrument
	InstrumentKindPCM = InstrumentKind(iota)
	// InstrumentKindOPL2 defines an OPL2 instrument
	InstrumentKindOPL2
)

// Instrument is an interface for instrument/sample data
type Instrument interface {
	IsInvalid() bool
	GetC2Spd() note.C2SPD
	SetC2Spd(note.C2SPD)
	GetDefaultVolume() volume.Volume
	GetID() InstrumentID
	GetSemitoneShift() int8
	SetFinetune(note.Finetune)
	GetFinetune() note.Finetune
	GetKind() InstrumentKind
	GetLength() sampling.Pos
	GetNewNoteAction() note.Action
	GetData() InstrumentDataIntf
	GetFilterFactory() FilterFactory
	GetAutoVibrato() voiceIntf.AutoVibrato
	IsReleaseNote(note.Note) bool
	IsStopNote(note.Note) bool
}
