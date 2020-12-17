// +build linux pulseaudio

package output

import (
	"gotracker/internal/output/pulseaudio"
	"gotracker/internal/player/render"
	"gotracker/internal/player/render/mixer"
)

type pulseaudioDevice struct {
	device
	mix mixer.Mixer
	pa  *pulseaudio.Client
}

func newPulseAudioDevice(settings Settings) (Device, error) {
	d := pulseaudioDevice{
		device: device{
			onRowOutput: settings.OnRowOutput,
		},
		mix: mixer.Mixer{
			Channels:      settings.Channels,
			BitsPerSample: settings.BitsPerSample,
		},
	}

	play, err := pulseaudio.New("Music", settings.SamplesPerSecond, settings.Channels, settings.BitsPerSample)
	if err != nil {
		return nil, err
	}

	d.pa = play
	return &d, nil
}

// Play starts the wave output device playing
func (d *pulseaudioDevice) Play(in <-chan render.RowRender) {
	panmixer := mixer.GetPanMixer(d.channels)
	for row := range in {
		mixedData := d.mix.Flatten(panmixer, row.SamplesLen, row.RenderData)
		d.pa.Output(mixedData)
		if d.onRowOutput != nil {
			d.onRowOutput(DeviceKindSoundCard, row)
		}
	}
}

// Close closes the wave output device
func (d *pulseaudioDevice) Close() {
	if d.pa != nil {
		d.pa.Close()
	}
}

func init() {
	deviceMap["pulseaudio"] = deviceDetails{
		create:   newPulseAudioDevice,
		kind:     DeviceKindSoundCard,
		priority: devicePriorityPulseAudio,
	}
}
