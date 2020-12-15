// +build windows

package output

import (
	"time"

	"gotracker/internal/output/win32/winmm"
	"gotracker/internal/player/render"
	"gotracker/internal/player/render/mixer"

	"github.com/pkg/errors"
)

type winmmDevice struct {
	device
	mix     mixer.Mixer
	waveout *winmm.WaveOut
}

func newWinMMDevice(settings Settings) (Device, error) {
	d := winmmDevice{
		mix: mixer.Mixer{
			Channels:      settings.Channels,
			BitsPerSample: settings.BitsPerSample,
		},
	}
	var err error
	d.waveout, err = winmm.New(settings.Channels, settings.SamplesPerSecond, settings.BitsPerSample)
	if err != nil {
		return nil, err
	}
	if d.waveout == nil {
		return nil, errors.New("could not create winmm device")
	}
	d.onRowOutput = settings.OnRowOutput
	return &d, nil
}

// Play starts the wave output device playing
func (d *winmmDevice) Play(in <-chan render.RowRender) {
	type RowWave struct {
		Wave *winmm.WaveOutData
		Row  render.RowRender
	}

	out := make(chan RowWave, 3)
	panmixer := mixer.GetPanMixer(d.mix.Channels)
	go func() {
		for row := range in {
			mixedData := d.mix.Flatten(panmixer, row.SamplesLen, row.RenderData)
			rowWave := RowWave{
				Wave: d.waveout.Write(mixedData),
				Row:  row,
			}
			out <- rowWave
		}
		close(out)
	}()
	for rowWave := range out {
		if d.onRowOutput != nil {
			d.onRowOutput(rowWave.Row)
		}
		for !d.waveout.IsHeaderFinished(rowWave.Wave) {
			time.Sleep(time.Microsecond * 1)
		}
	}
}

// Close closes the wave output device
func (d *winmmDevice) Close() {
	if d.waveout != nil {
		d.waveout.Close()
	}
}

func init() {
	deviceMap["winmm"] = deviceDetails{
		create:   newWinMMDevice,
		kind:     outputDeviceKindSoundCard,
		priority: outputDevicePriorityWinmm,
	}
}
