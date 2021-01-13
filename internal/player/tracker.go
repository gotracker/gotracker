package player

import (
	"errors"
	"time"

	"github.com/gotracker/gomixing/mixing"
	"github.com/gotracker/gomixing/panning"
	"github.com/gotracker/gomixing/volume"
	device "github.com/gotracker/gosound"

	"gotracker/internal/player/feature"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/render"
	"gotracker/internal/player/sampler"
)

// GetPremixDataIntf is an interface to getting the premix data from the tracker
type GetPremixDataIntf interface {
	GetPremixData() (*device.PremixData, error)
}

// Tracker is an extensible music tracker
type Tracker struct {
	intf.Playback
	render.OPL2Intf

	BaseClockRate float32
	Tickable      TickableIntf
	Premixable    GetPremixDataIntf

	s    *sampler.Sampler
	opl2 render.OPL2Chip

	globalVolume volume.Volume
	mixerVolume  volume.Volume

	panicOnUnknownEffect bool
}

// Update runs processing on the tracker, producing premixed sound data
func (t *Tracker) Update(deltaTime time.Duration, out chan<- *device.PremixData) error {
	premix, err := t.Generate(deltaTime)
	if err != nil {
		return err
	}
	if premix != nil && premix.Data != nil && len(premix.Data) != 0 {
		out <- premix
	}

	return nil
}

// Generate runs processing on the tracker, then returns the premixed sound data (if possible)
func (t *Tracker) Generate(deltaTime time.Duration) (*device.PremixData, error) {
	premix, err := t.renderTick()
	if err != nil {
		return nil, err
	}

	if premix != nil {
		if len(premix.Data) == 0 {
			cd := mixing.ChannelData{
				mixing.Data{
					Data:       nil,
					Pan:        panning.CenterAhead,
					Volume:     volume.Volume(0),
					SamplesLen: premix.SamplesLen,
				},
			}
			premix.Data = append(premix.Data, cd)
		}
		return premix, nil
	}

	return nil, nil
}

func (t *Tracker) renderTick() (*device.PremixData, error) {
	if err := DoTick(t.Tickable); err != nil {
		return nil, err
	}

	premix, err := t.Premixable.GetPremixData()
	if err != nil {
		return nil, err
	}

	if t.opl2 != nil {
		rr := [1]mixing.Data{}
		t.renderOPL2Tick(&rr[0],
			t.s.Mixer(),
			premix.SamplesLen)
		premix.Data = append(premix.Data, rr[:])

		// make room in the mixer for the OPL2 data
		// effectively, we can do this by calculating the new number (+1) of channels from the mixer volume (channels = reciprocal of mixer volume):
		//   numChannels = (1/mv) + 1
		// then by taking the reciprocal of it:
		//   1 / numChannels
		// but that ends up being simplified to:
		//   mv / (mv + 1)
		// and we get protection from div/0 in the process - provided, of course, that the mixerVolume is not exactly -1...
		mv := premix.MixerVolume
		premix.MixerVolume /= (mv + 1)
	}
	return premix, nil
}

func (t *Tracker) renderOPL2Tick(mixerData *mixing.Data, mix *mixing.Mixer, tickSamples int) {
	// make a stand-alone data buffer for this channel for this tick
	data := mix.NewMixBuffer(tickSamples)

	opl2data := make([]int32, tickSamples)

	if opl2 := t.opl2; opl2 != nil {
		opl2.GenerateBlock2(uint(tickSamples), opl2data)
	}

	for i, s := range opl2data {
		sv := volume.Volume(s) / 32768.0
		for c := range data {
			data[c][i] = sv
		}
	}
	*mixerData = mixing.Data{
		Data:       data,
		Pan:        panning.CenterAhead,
		Volume:     t.globalVolume,
		SamplesLen: tickSamples,
	}
}

// GetOPL2Chip returns the current song's OPL2 chip, if it's needed
func (t *Tracker) GetOPL2Chip() render.OPL2Chip {
	return t.opl2
}

// SetOPL2Chip sets the current song's OPL2 chip
func (t *Tracker) SetOPL2Chip(opl2 render.OPL2Chip) {
	t.opl2 = opl2
}

// SetupSampler configures the internal sampler
func (t *Tracker) SetupSampler(samplesPerSecond int, channels int, bitsPerSample int) error {
	t.s = sampler.NewSampler(samplesPerSecond, channels, bitsPerSample, t.BaseClockRate)
	if t.s == nil {
		return errors.New("NewSampler() returned nil")
	}

	return nil
}

// GetSampler returns the current sampler
func (t *Tracker) GetSampler() *sampler.Sampler {
	return t.s
}

// GetGlobalVolume returns the global volume value
func (t *Tracker) GetGlobalVolume() volume.Volume {
	return t.globalVolume
}

// SetGlobalVolume sets the global volume to the specified `vol` value
func (t *Tracker) SetGlobalVolume(vol volume.Volume) {
	t.globalVolume = vol
}

// GetMixerVolume returns the mixer volume value
func (t *Tracker) GetMixerVolume() volume.Volume {
	return t.mixerVolume
}

// SetMixerVolume sets the mixer volume to the specified `vol` value
func (t *Tracker) SetMixerVolume(vol volume.Volume) {
	t.mixerVolume = vol
}

// IgnoreUnknownEffect returns true if the tracker wants unknown effects to be ignored
func (t *Tracker) IgnoreUnknownEffect() bool {
	return !t.panicOnUnknownEffect
}

// DisableFeatures disables specified features
func (t *Tracker) DisableFeatures(features []feature.Feature) {
	for _, f := range features {
		switch f {
		case feature.IgnoreUnknownEffect:
			t.panicOnUnknownEffect = true
		}
	}
}
