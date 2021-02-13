package envelope

import (
	"gotracker/internal/loop"
	"gotracker/internal/player/intf"
)

// EnvPoint is a point for the envelope
type EnvPoint interface {
	Length() int
	Value(out interface{})
	Init(ticks int, value interface{})
}

// Envelope is an envelope for instruments
type Envelope struct {
	Enabled    bool
	Loop       loop.Loop
	Sustain    loop.Loop
	Values     []EnvPoint
	OnFinished func(intf.NoteControl)
}
