package intf

import (
	"gotracker/internal/player/note"

	"github.com/heucuva/gomixing/sampling"
	"github.com/heucuva/gomixing/volume"
)

// Instrument is an interface for instrument/sample data
type Instrument interface {
	sampling.SampleStream
	IsInvalid() bool
	GetC2Spd() note.C2SPD
	SetC2Spd(note.C2SPD)
	GetVolume() volume.Volume
	GetID() int
}
