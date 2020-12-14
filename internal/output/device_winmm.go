// +build windows

package output

import (
	"time"

	"gotracker/internal/output/winmm"
	"gotracker/internal/player/render"

	"github.com/pkg/errors"
)

type winmmDevice device

func newWinMMDevice(settings Settings) (Device, error) {
	d := winmmDevice{}
	var err error
	d.internal, err = winmm.New(settings.Channels, settings.SamplesPerSecond, settings.BitsPerSample)
	if err != nil {
		return nil, err
	}
	if d.internal == nil {
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

	hwo := *(d.internal.(*winmm.WaveOut))

	out := make(chan RowWave, 3)
	go func() {
		for row := range in {
			var rowWave RowWave
			rowWave.Wave = hwo.Write(row.RenderData)
			rowWave.Row = row
			out <- rowWave
		}
		close(out)
	}()
	for rowWave := range out {
		if d.onRowOutput != nil {
			d.onRowOutput(rowWave.Row)
		}
		for !hwo.IsHeaderFinished(rowWave.Wave) {
			time.Sleep(time.Microsecond * 1)
		}
	}
}

// Close closes the wave output device
func (d *winmmDevice) Close() {
	hwo := *(d.internal.(*winmm.WaveOut))
	hwo.Close()
}

func init() {
	deviceMap["winmm"] = deviceDetails{
		create:   newWinMMDevice,
		kind:     outputDeviceKindSoundCard,
		priority: outputDevicePriorityWinmm,
	}
}
