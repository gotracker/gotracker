package instrument

import (
	"time"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/intf"
)

// DataIntf is the interface to implementation-specific functions on an instrument
type DataIntf interface {
	GetSample(intf.NoteControl, sampling.Pos) volume.Matrix
	GetCurrentPanning(intf.NoteControl) panning.Position
	SetEnvelopePosition(intf.NoteControl, int)
	Initialize(intf.NoteControl) error
	Attack(intf.NoteControl)
	Release(intf.NoteControl)
	GetKeyOn(intf.NoteControl) bool
	Update(intf.NoteControl, time.Duration)
	UpdatePosition(intf.NoteControl, *sampling.Pos)
}
