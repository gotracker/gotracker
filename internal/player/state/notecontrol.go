package state

import (
	"time"

	"github.com/gotracker/gomixing/panning"

	"gotracker/internal/player/intf"
	voiceIntf "gotracker/internal/player/intf/voice"
	"gotracker/internal/player/note"
	"gotracker/internal/voice"
)

// NoteControl is an instance of the instrument on a particular output channel
type NoteControl struct {
	Voice  voiceIntf.Voice
	Output *intf.OutputChannel
}

// SetupVoice configures the voice using the instrument data interface provided
func (nc *NoteControl) SetupVoice(inst intf.Instrument) {
	nc.Voice = voice.New(inst, nc.Output)
}

// Clone clones the current note-control interface so that it doesn't collide with the existing one
func (nc *NoteControl) Clone() intf.NoteControl {
	c := *nc
	c.Voice = nc.Voice.Clone()

	return &c
}

// SetEnvelopePosition sets the envelope position(s) on the voice
func (nc *NoteControl) SetEnvelopePosition(pos int) {
	voiceIntf.SetVolumeEnvelopePosition(nc.Voice, pos)
	voiceIntf.SetPanEnvelopePosition(nc.Voice, pos)
	voiceIntf.SetPitchEnvelopePosition(nc.Voice, pos)
	voiceIntf.SetFilterEnvelopePosition(nc.Voice, pos)
}

// GetCurrentPeriodDelta returns the current pitch envelope value
func (nc *NoteControl) GetCurrentPeriodDelta() note.PeriodDelta {
	return voiceIntf.GetPeriodDelta(nc.Voice)
}

// GetCurrentFilterEnvValue returns the current filter envelope value
func (nc *NoteControl) GetCurrentFilterEnvValue() float32 {
	return voiceIntf.GetCurrentFilterEnvelope(nc.Voice)
}

// GetCurrentPanning returns the panning envelope position
func (nc *NoteControl) GetCurrentPanning() panning.Position {
	return voiceIntf.GetFinalPan(nc.Voice)
}

// GetOutputChannel returns the note-control's output channel
func (nc *NoteControl) GetOutputChannel() *intf.OutputChannel {
	return nc.Output
}

// GetVoice returns the voice that's on this instance
func (nc *NoteControl) GetVoice() voiceIntf.Voice {
	return nc.Voice
}

// Attack sets the key on flag for the instrument
func (nc *NoteControl) Attack() {
	if v := nc.Voice; v != nil {
		v.Attack()
	}
}

// Release clears the key on flag for the instrument
func (nc *NoteControl) Release() {
	if v := nc.Voice; v != nil {
		v.Release()
	}
}

// Fadeout sets the instrument to fading-out mode
func (nc *NoteControl) Fadeout() {
	if v := nc.Voice; v != nil {
		v.Fadeout()
	}
}

// GetKeyOn gets the key on flag for the instrument
func (nc *NoteControl) GetKeyOn() bool {
	if v := nc.Voice; v != nil {
		return v.IsKeyOn()
	}
	return false
}

// Update advances time by the amount specified by `tickDuration`
func (nc *NoteControl) Update(tickDuration time.Duration) {
	if v := nc.Voice; v != nil && nc.Output != nil {
		v.Advance(tickDuration)
	}
}

// IsVolumeEnvelopeEnabled returns true if the volume envelope is enabled
func (nc *NoteControl) IsVolumeEnvelopeEnabled() bool {
	return voiceIntf.IsVolumeEnvelopeEnabled(nc.Voice)
}

// IsDone returns true if the instrument has stopped
func (nc *NoteControl) IsDone() bool {
	return nc.Voice == nil || nc.Voice.IsDone()
}

// SetVolumeEnvelopeEnable sets the enable flag on the active volume envelope
func (nc *NoteControl) SetVolumeEnvelopeEnable(enabled bool) {
	voiceIntf.EnableVolumeEnvelope(nc.Voice, enabled)
}

// SetPanningEnvelopeEnable sets the enable flag on the active panning envelope
func (nc *NoteControl) SetPanningEnvelopeEnable(enabled bool) {
	voiceIntf.EnablePanEnvelope(nc.Voice, enabled)
}

// SetPitchEnvelopeEnable sets the enable flag on the active pitch/filter envelope
func (nc *NoteControl) SetPitchEnvelopeEnable(enabled bool) {
	voiceIntf.EnablePitchEnvelope(nc.Voice, enabled)
}
