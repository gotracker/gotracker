package state

import (
	"time"

	"github.com/gotracker/gomixing/mixing"
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/sampling"
	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/intf"
	"gotracker/internal/player/note"
)

// RenderState is the information needed to make an instrument play
type RenderState struct {
	Instrument   intf.Instrument
	Period       note.Period
	Volume       volume.Volume
	PeriodDelta  note.PeriodDelta
	volumeActive bool
	Pos          sampling.Pos
	Pan          panning.Position
}

// Reset sets the render state to defaults
func (r *RenderState) Reset() {
	r.Instrument = nil
	r.Period = nil
	r.Volume = 1
	r.PeriodDelta = 0
	r.volumeActive = true
	r.Pos = sampling.Pos{}
	r.Pan = panning.CenterAhead
}

// ActiveState is the active state of a channel
type ActiveState struct {
	RenderState
	NoteControl intf.NoteControl
}

// Reset sets the active state to defaults
func (a *ActiveState) Reset() {
	a.RenderState.Reset()
	a.NoteControl = nil
}

// Render renders an active channel's sample data for a the provided number of samples
func (a *ActiveState) Render(globalVolume volume.Volume, mix *mixing.Mixer, panmixer mixing.PanMixer, samplerSpeed float32, samples int, duration time.Duration) (*mixing.Data, error) {
	if a.Period == nil {
		return nil, nil
	}

	nc := a.NoteControl
	if nc == nil {
		return nil, nil
	}
	nc.SetVolume(a.Volume * globalVolume)
	period := a.Period.Add(a.PeriodDelta)
	nc.SetPeriod(period)

	samplerAdd := float32(period.GetSamplerAdd(float64(samplerSpeed)))

	nc.Update(duration)

	panning := nc.GetCurrentPanning()
	volMatrix := panmixer.GetMixingMatrix(panning)

	// make a stand-alone data buffer for this channel for this tick
	var data mixing.MixBuffer
	if a.volumeActive {
		data = mix.NewMixBuffer(samples)
		mixData := mixing.SampleMixIn{
			Sample:    sampling.NewSampler(nc, a.Pos, samplerAdd),
			StaticVol: volume.Volume(1.0),
			VolMatrix: volMatrix,
			MixPos:    0,
			MixLen:    samples,
		}
		data.MixInSample(mixData)
	}

	a.Pos.Add(samplerAdd * float32(samples))

	return &mixing.Data{
		Data:       data,
		Pan:        a.Pan,
		Volume:     volume.Volume(1.0),
		SamplesLen: samples,
	}, nil
}
