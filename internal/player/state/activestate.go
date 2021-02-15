package state

import (
	"time"

	"github.com/gotracker/gomixing/mixing"
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"

	panutil "gotracker/internal/pan"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/intf/voice"
	"gotracker/internal/player/note"
)

// ActiveState is the active state of a channel
type ActiveState struct {
	intf.PlaybackState
	VoiceActive bool
	Enabled     bool
	NoteControl intf.NoteControl
	PeriodDelta note.PeriodDelta
}

// Reset sets the active state to defaults
func (a *ActiveState) Reset() {
	a.PlaybackState.Reset()
	a.VoiceActive = true
	a.Enabled = true
	a.NoteControl = nil
	a.PeriodDelta = 0
}

// Clone clones the active state so that various interfaces do not collide
func (a *ActiveState) Clone() ActiveState {
	var c ActiveState = *a
	if a.NoteControl != nil {
		c.NoteControl = a.NoteControl.Clone()
	}

	return c
}

// RenderStatesTogether renders a channel's series of sample data for a the provided number of samples
func RenderStatesTogether(states []*ActiveState, mix *mixing.Mixer, panmixer mixing.PanMixer, samplerSpeed float32, samples int, duration time.Duration) (*mixing.Data, []*ActiveState) {
	data := mix.NewMixBuffer(samples)
	var firstPan *panning.Position
	var tempPan panning.Position

	participatingStates := []*ActiveState{}
	for _, a := range states {
		if !a.Enabled || a.Period == nil {
			continue
		}

		nc := a.NoteControl
		if nc == nil || nc.IsDone() {
			continue
		}

		ncv := nc.GetVoice()
		if ncv == nil {
			continue
		}

		// Commit the playback settings to the note-control
		voice.SetPeriod(ncv, a.Period)
		voice.SetVolume(ncv, a.Volume)
		voice.SetPos(ncv, a.Pos)
		voice.SetPan(ncv, a.Pan)

		if firstPan == nil {
			tempPan = voice.GetPan(ncv)
			firstPan = &tempPan
		}

		voice.SetPeriodDelta(ncv, a.PeriodDelta)

		// the period might be updated by the auto-vibrato system, here
		nc.Update(duration)

		// ... so grab the new value now.
		period := voice.GetFinalPeriod(ncv)
		panVoice := voice.GetFinalPan(ncv)

		samplerAdd := float32(period.GetSamplerAdd(float64(samplerSpeed)))

		panOrig := nc.GetCurrentPanning()
		panDiff := panutil.GetPanningDifference(*firstPan, panOrig)
		panning := panutil.CalculateCombinedPanning(panVoice, panDiff)
		volMatrix := panmixer.GetMixingMatrix(panning)

		// make a stand-alone data buffer for this channel for this tick
		if a.VoiceActive {
			mixData := mixing.SampleMixIn{
				Sample:    ncv.GetSampler(samplerSpeed, nc.GetOutputChannel()),
				StaticVol: volume.Volume(1.0),
				VolMatrix: volMatrix,
				MixPos:    0,
				MixLen:    samples,
			}
			if mixData.Sample != nil {
				data.MixInSample(mixData)
			}
		}

		a.Pos = voice.GetPos(ncv)
		a.Pos.Add(samplerAdd * float32(samples))

		participatingStates = append(participatingStates, a)
	}

	if firstPan == nil {
		return nil, nil
	}

	mixData := mixing.Data{
		Data:       data,
		Pan:        *firstPan,
		Volume:     volume.Volume(1.0),
		SamplesLen: samples,
	}

	return &mixData, participatingStates
}
