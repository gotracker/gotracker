package state

import (
	"time"

	"github.com/gotracker/gomixing/mixing"
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"
	"github.com/gotracker/voice"

	"gotracker/internal/player/note"
)

// Active is the active state of a channel
type Active struct {
	Playback
	Voice       voice.Voice
	PeriodDelta note.PeriodDelta
}

// Reset sets the active state to defaults
func (a *Active) Reset() {
	a.Playback.Reset()
	a.Voice = nil
	a.PeriodDelta = 0
}

// Clone clones the active state so that various interfaces do not collide
func (a *Active) Clone() Active {
	var c Active = *a
	if a.Voice != nil {
		c.Voice = a.Voice.Clone()
	}

	return c
}

// RenderStatesTogether renders a channel's series of sample data for a the provided number of samples
func RenderStatesTogether(states []*Active, mix *mixing.Mixer, panmixer mixing.PanMixer, samplerSpeed float32, samples int, duration time.Duration) ([]mixing.Data, []*Active) {
	data := mix.NewMixBuffer(samples)

	mixData := []mixing.Data{}

	centerAheadPan := panmixer.GetMixingMatrix(panning.CenterAhead)

	participatingStates := []*Active{}
	for _, a := range states {
		if a.Period == nil {
			continue
		}

		ncv := a.Voice
		if ncv == nil || ncv.IsDone() {
			continue
		}

		// Commit the playback settings to the note-control
		voice.SetPeriod(ncv, a.Period)
		voice.SetVolume(ncv, a.Volume)
		voice.SetPos(ncv, a.Pos)
		voice.SetPan(ncv, a.Pan)

		voice.SetPeriodDelta(ncv, a.PeriodDelta)

		// the period might be updated by the auto-vibrato system, here
		ncv.Advance(duration)

		// ... so grab the new value now.
		period := voice.GetFinalPeriod(ncv)
		pan := voice.GetFinalPan(ncv)

		samplerAdd := float32(period.GetSamplerAdd(float64(samplerSpeed)))

		// make a stand-alone data buffer for this channel for this tick
		if ncv.IsActive() {
			sampleData := mixing.SampleMixIn{
				Sample:    ncv.GetSampler(samplerSpeed),
				StaticVol: volume.Volume(1.0),
				VolMatrix: centerAheadPan,
				MixPos:    0,
				MixLen:    samples,
			}
			if sampleData.Sample != nil {
				data.MixInSample(sampleData)
			}
		}

		mixData = append(mixData, mixing.Data{
			Data:       data,
			Pan:        pan,
			Volume:     volume.Volume(1.0),
			SamplesLen: samples,
		})

		a.Pos = voice.GetPos(ncv)
		a.Pos.Add(samplerAdd * float32(samples))

		participatingStates = append(participatingStates, a)
	}

	return mixData, participatingStates
}
