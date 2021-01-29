package intf

import (
	"time"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/oscillator"
	"gotracker/internal/player/note"
)

// NoteControl is an interface for an instrument on a particular output channel
type NoteControl interface {
	sampling.SampleStream

	Clone() NoteControl
	GetOutputChannel() *OutputChannel
	GetCurrentPeriodDelta() note.PeriodDelta
	GetCurrentFilterEnvValue() float32
	GetCurrentPanning() panning.Position
	Attack()
	Release()
	Fadeout()
	GetKeyOn() bool
	Update(time.Duration)
	SetFilter(Filter)
	SetData(interface{})
	GetData() interface{}
	SetEnvelopePosition(int)
	GetPlaybackState() *PlaybackState
	GetAutoVibratoState() *AutoVibratoState
	IsVolumeEnvelopeEnabled() bool
	IsDone() bool
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

// AutoVibratoState is the information needed to make an instrument auto-vibrato
type AutoVibratoState struct {
	Osc   oscillator.Oscillator
	Ticks int
}

// Reset sets the auto-vibrato state to defaults
func (av *AutoVibratoState) Reset() {
	if av.Osc != nil {
		av.Osc.Reset()
	}
	av.Ticks = 0
}
