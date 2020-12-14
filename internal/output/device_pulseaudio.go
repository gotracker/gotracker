// +build pulseaudio

package output

import (
	"gotracker/internal/output/pulseaudio"
	"gotracker/internal/player/render"
	"gotracker/internal/player/render/mixer"
)

type pulseaudioDevice struct {
	device
	channels      int
	bitsPerSample int
	mix           mixer.Mixer
}

func newPulseAudioDevice(settings Settings) (Device, error) {
	d := pulseaudioDevice{
		channels:      settings.Channels,
		bitsPerSample: settings.BitsPerSample,
	}

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
	play := d.internal.(*pulseaudio.Client)

	panmixer := mixer.GetPanMixer(d.channels)
	for row := range in {
		data := d.mix.NewMixBuffer(d.channels, row.SamplesLen)
		for _, rdata := range row.RenderData {
			pos := 0
			for _, cdata := range rdata {
				if cdata.Flush != nil {
					cdata.Flush()
				}
				if len(cdata.Data) > 0 {
					volMtx := cdata.Volume.Apply(panmixer.GetMixingMatrix(cdata.Pan)...)
					data.Add(pos, cdata.Data, volMtx)
				}
				pos += cdata.SamplesLen
			}
		}
		mixedData := data.ToRenderData(row.SamplesLen, d.bitsPerSample, len(row.RenderData))
		play.Output(mixedData)
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
	deviceMap["pulseaudio"] = deviceDetails{
		create:   newPulseAudioDevice,
		kind:     outputDeviceKindSoundCard,
		priority: outputDevicePriorityPulseAudio,
	}
}
