package intf

import (
	"time"

	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/note"
)

// InstrumentID is an identifier for an instrument/sample that means something to the format
type InstrumentID interface {
	IsEmpty() bool
}

// Instrument is an interface for instrument/sample data
type Instrument interface {
	IsInvalid() bool
	GetC2Spd() note.C2SPD
	SetC2Spd(note.C2SPD)
	GetDefaultVolume() volume.Volume
	GetID() InstrumentID
	GetSemitoneShift() int8
	InstantiateOnChannel(int, Filter) NoteControl
	SetFinetune(int8)
	GetFinetune() int8

	GetSample(NoteControl, sampling.Pos) volume.Matrix
	Attack(NoteControl)
	Release(NoteControl)
	NoteCut(NoteControl)
	GetKeyOn(NoteControl) bool
	Update(NoteControl, time.Duration)
}
