package intf

import "gotracker/internal/player/volume"

// Instrument is an interface for instrument/sample data
type Instrument interface {
	IsInvalid() bool
	GetC2Spd() uint16
	SetC2Spd(uint16)
	GetVolume() volume.Volume
	IsLooped() bool
	GetLoopBegin() int
	GetLoopEnd() int
	GetSample(int) volume.Volume
	GetLength() int
	GetId() int
}
