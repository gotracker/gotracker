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

// ChannelFilterFactory is a function type that builds a filter with an input parameter taking a value between 0 and 1
type ChannelFilterFactory func(float32) Filter

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
	GetKind() note.InstrumentKind
	GetLength() sampling.Pos
	GetNewNoteAction() note.Action
	GetData() InstrumentDataIntf
	GetChannelFilterFactory() ChannelFilterFactory
	GetAutoVibrato() voiceIntf.AutoVibrato
}
