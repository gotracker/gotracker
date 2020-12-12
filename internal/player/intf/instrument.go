package intf

import (
	"gotracker/internal/player/note"
	"gotracker/internal/player/volume"
)

// Instrument is an interface for instrument/sample data
type Instrument interface {
	IsInvalid() bool
	GetC2Spd() note.C2SPD
	SetC2Spd(note.C2SPD)
	GetVolume() volume.Volume
	IsLooped() bool
	GetLoopBegin() float32
	GetLoopEnd() float32
	GetSample(float32) volume.VolumeMatrix
	GetLength() float32
	GetID() int
}
