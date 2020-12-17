package intf

import (
	"github.com/heucuva/gomixing/sampling"
	"github.com/heucuva/gomixing/volume"

	"gotracker/internal/player/note"
)

// Instrument is an interface for instrument/sample data
type Instrument interface {
	sampling.SampleStream
	IsInvalid() bool
	GetC2Spd() note.C2SPD
	SetC2Spd(note.C2SPD)
	GetVolume() volume.Volume
	IsLooped() bool
	GetLoopBegin() sampling.Pos
	GetLoopEnd() sampling.Pos
	GetLength() sampling.Pos
	GetID() int
}
