// +build windows

package device

import (
	"errors"
	"time"

	"gotracker/internal/audio/mixing"
	"gotracker/internal/output/device/win32/winmm"
)

type winmmDevice struct {
	device
	mix     mixing.Mixer
	waveout *winmm.WaveOut
}

func newWinMMDevice(settings Settings) (Device, error) {
	d := winmmDevice{
		device: device{
			onRowOutput: settings.OnRowOutput,
		},
		mix: mixing.Mixer{
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
	return &d, nil
}

// Play starts the wave output device playing
func (d *winmmDevice) Play(in <-chan *PremixData) {
	type RowWave struct {
		Wave *winmm.WaveOutData
		Row  *PremixData
	}

	out := make(chan RowWave, 3)
	panmixer := mixing.GetPanMixer(d.mix.Channels)
	go func() {
		for row := range in {
			mixedData := d.mix.Flatten(panmixer, row.SamplesLen, row.Data)
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
			d.onRowOutput(KindSoundCard, rowWave.Row)
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
	Map["winmm"] = deviceDetails{
		create: newWinMMDevice,
		kind:   KindSoundCard,
	}
}
