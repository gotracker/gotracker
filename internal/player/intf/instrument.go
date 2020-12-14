package intf

import (
	"gotracker/internal/player/note"
	"gotracker/internal/player/sample"
	"gotracker/internal/player/volume"
)

// Instrument is an interface for instrument/sample data
type Instrument interface {
	IsInvalid() bool
	GetC2Spd() note.C2SPD
	SetC2Spd(note.C2SPD)
	GetVolume() volume.Volume
	IsLooped() bool
	GetLoopBegin() sample.Pos
	GetLoopEnd() sample.Pos
	GetSample(sample.Pos) volume.VolumeMatrix
	GetLength() sample.Pos
	GetID() int
}
