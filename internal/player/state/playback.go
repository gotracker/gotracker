package state

import (
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/song"
	"gotracker/internal/song/note"
)

// Playback is the information needed to make an instrument play
type Playback struct {
	Instrument song.Instrument
	Period     note.Period
	Volume     volume.Volume
	Pos        sampling.Pos
	Pan        panning.Position
}

// Reset sets the render state to defaults
func (p *Playback) Reset() {
	p.Instrument = nil
	p.Period = nil
	p.Volume = 1
	p.Pos = sampling.Pos{}
	p.Pan = panning.CenterAhead
}
