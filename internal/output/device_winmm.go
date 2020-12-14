// +build windows

package output

import (
	"time"

	"gotracker/internal/output/winmm"
	"gotracker/internal/player/render"
	"gotracker/internal/player/render/mixer"

	"github.com/pkg/errors"
)

type winmmDevice struct {
	device
	channels      int
	bitsPerSample int
	mix           mixer.Mixer
}

func newWinMMDevice(settings Settings) (Device, error) {
	d := winmmDevice{
		channels:      settings.Channels,
		bitsPerSample: settings.BitsPerSample,
	}
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

	hwo := d.internal.(*winmm.WaveOut)

	out := make(chan RowWave, 3)
	panmixer := mixer.GetPanMixer(d.channels)
	go func() {
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
			rowWave := RowWave{
				Wave: hwo.Write(mixedData),
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
		for !hwo.IsHeaderFinished(rowWave.Wave) {
			time.Sleep(time.Microsecond * 1)
		}
	}
}

// Close closes the wave output device
func (d *winmmDevice) Close() {
	hwo := d.internal.(*winmm.WaveOut)
	hwo.Close()
}

func init() {
	deviceMap["winmm"] = deviceDetails{
		create:   newWinMMDevice,
		kind:     outputDeviceKindSoundCard,
		priority: outputDevicePriorityWinmm,
	}
}
