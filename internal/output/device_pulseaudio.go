// +build pulseaudio

package output

import (
	"gotracker/internal/output/pulseaudio"
	"gotracker/internal/player/render"
)

type pulseaudioDevice device

func newPulseAudioDevice(settings Settings) (Device, error) {
	d := pulseaudioDevice{}

	play, err := pulseaudio.New("Music", settings.SamplesPerSecond, settings.Channels, settings.BitsPerSample)
	if err != nil {
		return nil, err
	}

	d.internal = play
	d.onRowOutput = settings.OnRowOutput
	return &d, nil
}

// Play starts the wave output device playing
func (d *pulseaudioDevice) Play(in <-chan render.RowRender) {
	play := *(d.internal.(*pulseaudio.Client))

	for row := range in {
		play.Output(row.RenderData)
		if d.onRowOutput != nil {
			d.onRowOutput(row)
		}
	}
}

// Close closes the wave output device
func (d *pulseaudioDevice) Close() {
	play := *(d.internal.(*pulseaudio.Client))
	play.Close()
}

func init() {
	deviceMap["pulseaudio"] = newPulseAudioDevice
	DefaultOutputDeviceName = "pulseaudio"
}
