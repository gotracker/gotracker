package state

import (
	"time"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// playbackState is the information needed to make an instrument play
type playbackState struct {
	Instrument intf.Instrument
	Period     note.Period
	Volume     volume.Volume
	Pos        sampling.Pos
	Pan        panning.Position
}

// Reset sets the render state to defaults
func (p *playbackState) Reset() {
	p.Instrument = nil
	p.Period = nil
	p.Volume = 1
	p.Pos = sampling.Pos{}
	p.Pan = panning.CenterAhead
}

// NoteControl is an instance of the instrument on a particular output channel
type NoteControl struct {
	intf.NoteControl
	playbackState

	Data   interface{}
	Output *intf.OutputChannel
}

// GetSample returns the sample at position `pos` in the instrument
func (nc *NoteControl) GetSample(pos sampling.Pos) volume.Matrix {
	if inst := nc.Instrument; inst != nil {
		dry := inst.GetSample(nc, pos)
		if nc.Output != nil {
			return nc.Output.ApplyFilter(dry)
		}
	}
	return nil
}

// GetCurrentPanning returns the panning envelope position
func (nc *NoteControl) GetCurrentPanning() panning.Position {
	if inst := nc.Instrument; inst != nil {
		return inst.GetCurrentPanning(nc)
	}
	return panning.CenterAhead
}

// GetOutputChannel returns the note-control's output channel
func (nc *NoteControl) GetOutputChannel() *intf.OutputChannel {
	return nc.Output
}

// GetInstrument returns the instrument that's on this instance
func (nc *NoteControl) GetInstrument() intf.Instrument {
	return nc.Instrument
}

// Attack sets the key on flag for the instrument
func (nc *NoteControl) Attack() {
	if inst := nc.Instrument; inst != nil {
		inst.Attack(nc)
	}
}

// Release clears the key on flag for the instrument
func (nc *NoteControl) Release() {
	if inst := nc.Instrument; inst != nil {
		inst.Release(nc)
	}
}

// GetKeyOn gets the key on flag for the instrument
func (nc *NoteControl) GetKeyOn() bool {
	if inst := nc.Instrument; inst != nil {
		return inst.GetKeyOn(nc)
	}
	return false
}

// Update advances time by the amount specified by `tickDuration`
func (nc *NoteControl) Update(tickDuration time.Duration) {
	if inst := nc.Instrument; inst != nil {
		inst.Update(nc, tickDuration)
	}
}

// SetVolume sets the active note-control's volume
func (nc *NoteControl) SetVolume(vol volume.Volume) {
	nc.Volume = vol
}

// GetVolume sets the active note-control's volume
func (nc *NoteControl) GetVolume() volume.Volume {
	return nc.Volume
}

// SetPeriod sets the active note-control's period
func (nc *NoteControl) SetPeriod(period note.Period) {
	nc.Period = period
}

// GetPeriod gets the active note-control's period
func (nc *NoteControl) GetPeriod() note.Period {
	return nc.Period
}

// SetData sets the data interface for the note-control
func (nc *NoteControl) SetData(data interface{}) {
	nc.Data = data
}

// GetData gets the data interface for the note-control
func (nc *NoteControl) GetData() interface{} {
	return nc.Data
}
