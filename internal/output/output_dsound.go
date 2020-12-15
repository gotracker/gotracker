// +build windows,dsound

package output

import (
	"gotracker/internal/output/win32"
	"gotracker/internal/output/win32/dsound"
	"gotracker/internal/player/render"
	"gotracker/internal/player/render/mixer"
	"time"

	"github.com/pkg/errors"
)

type dsoundDevice struct {
	device

	ds           *dsound.DirectSound
	lpdsbPrimary *dsound.Buffer
	wfx          *win32.WAVEFORMATEX

	mix mixer.Mixer
}

func newDSoundDevice(settings Settings) (Device, error) {
	d := dsoundDevice{
		mix: mixer.Mixer{
			Channels:      settings.Channels,
			BitsPerSample: settings.BitsPerSample,
		},
	}
	preferredDeviceName := ""

	ds, err := dsound.NewDSound(preferredDeviceName)
	if err != nil {
		return nil, err
	}
	d.ds = ds
	if d.ds == nil {
		return nil, errors.New("could not create dsound device")
	}

	lpdsbPrimary, wfx, err := ds.CreateSoundBufferPrimary(settings.Channels, settings.SamplesPerSecond, settings.BitsPerSample)
	if err != nil {
		ds.Close()
		return nil, err
	}
	d.lpdsbPrimary = lpdsbPrimary
	d.wfx = wfx

	d.onRowOutput = settings.OnRowOutput
	return &d, nil
}

// Play starts the wave output device playing
func (d *dsoundDevice) Play(in <-chan render.RowRender) {
	type RowWave struct {
		LpDsb *dsound.Buffer
		Row   render.RowRender
	}

	out := make(chan RowWave, 3)
	panmixer := mixer.GetPanMixer(d.mix.Channels)
	go func() {
		for row := range in {
			var rowWave RowWave
			numBytes := row.SamplesLen * int(d.wfx.NBlockAlign)
			lpdsb, err := d.ds.CreateSoundBufferSecondary(d.wfx, numBytes)
			if err != nil {
				continue
			}
			segments, err := lpdsb.Lock(0, numBytes)
			if err != nil {
				lpdsb.Release()
				continue
			}
			d.mix.FlattenTo(segments[0], panmixer, row.SamplesLen, row.RenderData)
			if err := lpdsb.Unlock(segments); err != nil {
				lpdsb.Release()
				continue
			}
			rowWave.LpDsb = lpdsb
			rowWave.Row = row
			out <- rowWave
		}
		close(out)
	}()
	for rowWave := range out {
		rowWave.LpDsb.Play(false)
		if d.onRowOutput != nil {
			d.onRowOutput(rowWave.Row)
		}
		for {
			status, err := rowWave.LpDsb.GetStatus()
			if err != nil {
				break
			}
			if (status & win32.DSBSTATUS_PLAYING) == 0 {
				break
			}
			time.Sleep(time.Microsecond * 1)
		}
		rowWave.LpDsb.Release()
	}
}

// Close closes the wave output device
func (d *dsoundDevice) Close() {
	if d.lpdsbPrimary != nil {
		d.lpdsbPrimary.Release()
	}
	if d.ds != nil {
		d.ds.Close()
	}
}

func init() {
	deviceMap["dsound"] = deviceDetails{
		create:   newDSoundDevice,
		kind:     outputDeviceKindSoundCard,
		priority: outputDevicePriorityDSound,
	}
}
