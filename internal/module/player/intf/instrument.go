package intf

import (
	"gotracker/internal/audio/sampling"
	"gotracker/internal/audio/volume"
	"gotracker/internal/module/player/note"
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
