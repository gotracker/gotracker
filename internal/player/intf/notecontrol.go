package intf

import (
	"gotracker/internal/player/note"
	"time"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"
)

// NoteControl is an interface for an instrument on a particular output channel
type NoteControl interface {
	sampling.SampleStream

	GetOutputChannel() *OutputChannel
	GetCurrentPanning() panning.Position
	Attack()
	Release()
	GetKeyOn() bool
	Update(time.Duration)
	SetFilter(Filter)
	SetData(interface{})
	GetData() interface{}
	SetEnvelopePosition(int)
	GetPlaybackState() *PlaybackState
}

// PlaybackState is the information needed to make an instrument play
type PlaybackState struct {
	Instrument Instrument
	Period     note.Period
	Volume     volume.Volume
	Pos        sampling.Pos
	Pan        panning.Position
}

// Reset sets the render state to defaults
func (p *PlaybackState) Reset() {
	p.Instrument = nil
	p.Period = nil
	p.Volume = 1
	p.Pos = sampling.Pos{}
	p.Pan = panning.CenterAhead
}
