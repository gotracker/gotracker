package intf

import (
	"gotracker/internal/player/note"
	"time"

	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"
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
	GetVolume() volume.Volume
	GetID() InstrumentID
	GetSemitoneShift() int8
	InstantiateOnChannel(int, Filter) InstrumentOnChannel
}

// InstrumentOnChannel is an interface for an instrument on a particular output channel
type InstrumentOnChannel interface {
	sampling.SampleStream

	GetInstrument() Instrument
	SetKeyOn(note.Period, bool)
	GetKeyOn() bool
	Update(time.Duration)
	SetFilter(Filter)
	SetVolume(volume.Volume)
	SetPeriod(note.Period)
}
