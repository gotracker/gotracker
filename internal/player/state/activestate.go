package state

import (
	"time"

	"github.com/gotracker/gomixing/mixing"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// activeState is the active state of a channel
type activeState struct {
	intf.PlaybackState
	VoiceActive bool
	NoteControl intf.NoteControl
	PeriodDelta note.PeriodDelta
}

// Reset sets the active state to defaults
func (a *activeState) Reset() {
	a.PlaybackState.Reset()
	a.VoiceActive = true
	a.NoteControl = nil
	a.PeriodDelta = 0
}

// Render renders an active channel's sample data for a the provided number of samples
func (a *activeState) Render(globalVolume volume.Volume, mix *mixing.Mixer, panmixer mixing.PanMixer, samplerSpeed float32, samples int, duration time.Duration) (*mixing.Data, error) {
	if a.Period == nil {
		return nil, nil
	}

	nc := a.NoteControl
	if nc == nil {
		return nil, nil
	}
	ncs := nc.GetPlaybackState()
	if ncs == nil {
		return nil, nil
	}

	*ncs = a.PlaybackState

	ncs.Volume *= globalVolume
	period := ncs.Period.Add(a.PeriodDelta)
	ncs.Period = period

	samplerAdd := float32(period.GetSamplerAdd(float64(samplerSpeed)))

	nc.Update(duration)

	panning := nc.GetCurrentPanning()
	volMatrix := panmixer.GetMixingMatrix(panning)

	// make a stand-alone data buffer for this channel for this tick
	var data mixing.MixBuffer
	if a.VoiceActive {
		data = mix.NewMixBuffer(samples)
		mixData := mixing.SampleMixIn{
			Sample:    sampling.NewSampler(nc, ncs.Pos, samplerAdd),
			StaticVol: volume.Volume(1.0),
			VolMatrix: volMatrix,
			MixPos:    0,
			MixLen:    samples,
		}
		data.MixInSample(mixData)
	}

	a.Pos = ncs.Pos
	a.Pos.Add(samplerAdd * float32(samples))

	return &mixing.Data{
		Data:       data,
		Pan:        ncs.Pan,
		Volume:     volume.Volume(1.0),
		SamplesLen: samples,
	}, nil
}
