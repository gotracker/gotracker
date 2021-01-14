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
	Instrument  intf.Instrument
	Period      note.Period
	Volume      volume.Volume
	VoiceActive bool
	Pos         sampling.Pos
	Pan         panning.Position
}

// Reset sets the render state to defaults
func (p *playbackState) Reset() {
	p.Instrument = nil
	p.Period = nil
	p.Volume = 1
	p.VoiceActive = true
	p.Pos = sampling.Pos{}
	p.Pan = panning.CenterAhead
}

// NoteControl is an instance of the instrument on a particular output channel
type NoteControl struct {
	intf.NoteControl
	playbackState

	OutputChannelNum int
	Data             interface{}
	Filter           intf.Filter
	Playback         intf.Playback
}

// GetSample returns the sample at position `pos` in the instrument
func (nc *NoteControl) GetSample(pos sampling.Pos) volume.Matrix {
	if inst := nc.Instrument; inst != nil {
		dry := inst.GetSample(nc, pos)
		if nc.Filter != nil {
			wet := nc.Filter.Filter(dry)
			return wet
		}
		return dry
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

// GetOutputChannelNum returns the note-control's output channel number
func (nc *NoteControl) GetOutputChannelNum() int {
	return nc.OutputChannelNum
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

// NoteCut cuts the current playback of the instrument
func (nc *NoteControl) NoteCut() {
	if inst := nc.Instrument; inst != nil {
		inst.NoteCut(nc)
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

// SetFilter sets the active filter on the instrument (which should be the same as what's on the channel)
func (nc *NoteControl) SetFilter(filter intf.Filter) {
	nc.Filter = filter
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

// SetPlayback sets the playback interface for the note-control
func (nc *NoteControl) SetPlayback(pb intf.Playback) {
	nc.Playback = pb
}

// GetPlayback gets the playback interface for the note-control
func (nc *NoteControl) GetPlayback() intf.Playback {
	return nc.Playback
}

// SetData sets the data interface for the note-control
func (nc *NoteControl) SetData(data interface{}) {
	nc.Data = data
}

// GetData gets the data interface for the note-control
func (nc *NoteControl) GetData() interface{} {
	return nc.Data
}
