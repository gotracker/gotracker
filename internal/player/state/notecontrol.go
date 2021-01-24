package state

import (
	"time"

	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// NoteControl is an instance of the instrument on a particular output channel
type NoteControl struct {
	intf.NoteControl
	intf.PlaybackState
	intf.AutoVibratoState

	Data   interface{}
	Output *intf.OutputChannel
}

// Clone clones the current note-control interface so that it doesn't collide with the existing one
func (nc *NoteControl) Clone() intf.NoteControl {
	c := *nc
	if inst := c.Instrument; inst != nil {
		c.Data = inst.CloneData(&c)
	}

	return &c
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

// GetCurrentPeriodDelta returns the current pitch envelope value
func (nc *NoteControl) GetCurrentPeriodDelta() note.PeriodDelta {
	if inst := nc.Instrument; inst != nil {
		return inst.GetCurrentPeriodDelta(nc)
	}
	return note.PeriodDelta(0)
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
		nc.AutoVibratoState.Reset()
		inst.Attack(nc)
	}
}

// Release clears the key on flag for the instrument
func (nc *NoteControl) Release() {
	if inst := nc.Instrument; inst != nil {
		inst.Release(nc)
	}
}

// Fadeout sets the instrument to fading-out mode
func (nc *NoteControl) Fadeout() {
	if inst := nc.Instrument; inst != nil {
		inst.Fadeout(nc)
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

// SetData sets the data interface for the note-control
func (nc *NoteControl) SetData(data interface{}) {
	nc.Data = data
}

// GetData gets the data interface for the note-control
func (nc *NoteControl) GetData() interface{} {
	return nc.Data
}

// GetPlaybackState returns the current, mutable playback state
func (nc *NoteControl) GetPlaybackState() *intf.PlaybackState {
	return &nc.PlaybackState
}

// GetAutoVibratoState returns the current, mutable auto-vibrato state
func (nc *NoteControl) GetAutoVibratoState() *intf.AutoVibratoState {
	return &nc.AutoVibratoState
}

// IsVolumeEnvelopeEnabled returns true if the volume envelope is enabled
func (nc *NoteControl) IsVolumeEnvelopeEnabled() bool {
	if inst := nc.Instrument; inst != nil {
		return inst.IsVolumeEnvelopeEnabled()
	}
	return false
}
