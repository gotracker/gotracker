package intf

import (
	"gotracker/internal/player/note"
	"time"

	"github.com/heucuva/gomixing/sampling"
	"github.com/heucuva/gomixing/volume"
)

// Instrument is an interface for instrument/sample data
type Instrument interface {
	IsInvalid() bool
	GetC2Spd() note.C2SPD
	SetC2Spd(note.C2SPD)
	GetVolume() volume.Volume
	GetID() int
	InstantiateOnChannel(int, Filter) InstrumentOnChannel
}

// InstrumentOnChannel is an interface for an instrument on a particular output channel
type InstrumentOnChannel interface {
	sampling.SampleStream

	GetInstrument() Instrument
	SetKeyOn(note.Semitone, bool)
	GetKeyOn() bool
	Update(time.Duration)
	SetFilter(Filter)
}
