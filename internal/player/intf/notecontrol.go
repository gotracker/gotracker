package intf

import (
	"time"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/note"
)

// NoteControl is an interface for an instrument on a particular output channel
type NoteControl interface {
	sampling.SampleStream

	GetOutputChannelNum() int
	GetInstrument() Instrument
	GetCurrentPanning() panning.Position
	Attack()
	Release()
	NoteCut()
	GetKeyOn() bool
	Update(time.Duration)
	SetFilter(Filter)
	SetVolume(volume.Volume)
	GetVolume() volume.Volume
	SetPeriod(note.Period)
	GetPeriod() note.Period
	SetPlayback(Playback)
	GetPlayback() Playback
	SetData(interface{})
	GetData() interface{}
	SetEnvelopePosition(int)
}
