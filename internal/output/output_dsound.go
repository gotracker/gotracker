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
		PlayOffset uint32
		Row        render.RowRender
	}

	event, err := win32.CreateEvent()
	if err != nil {
		return
	}
	defer win32.CloseHandle(event)

	out := make(chan RowWave, 3)
	panmixer := mixer.GetPanMixer(d.mix.Channels)

	playbackSize := int(d.wfx.NAvgBytesPerSec * 2)
	lpdsb, err := d.ds.CreateSoundBufferSecondary(d.wfx, playbackSize)
	if err != nil {
		return
	}
	defer lpdsb.Release()

	notify, err := lpdsb.GetNotify()
	if err != nil {
		return
	}
	defer notify.Release()

	// play (looping)
	lpdsb.Play(true)

	go func() {
		writePos := 0
		for row := range in {
			var rowWave RowWave
			//_, writePos, err := lpdsb.GetCurrentPosition()
			numBytes := row.SamplesLen * int(d.wfx.NBlockAlign)
			segments, err := lpdsb.Lock(writePos, numBytes)
			if err != nil {
				continue
			}
			d.mix.FlattenTo(segments, panmixer, row.SamplesLen, row.RenderData)
			if err := lpdsb.Unlock(segments); err != nil {
				continue
			}
			rowWave.Row = row
			writePos = (writePos + numBytes) % playbackSize
			rowWave.PlayOffset = uint32(writePos)
			out <- rowWave
		}
		close(out)
	}()
	prevPlay := uint32(0)
	for rowWave := range out {
		for {
			playPos, _, _ := lpdsb.GetCurrentPosition()
			triggered := false
			if prevPlay <= playPos {
				if playPos >= rowWave.PlayOffset {
					triggered = true
				}
			} else {
				if playPos <= rowWave.PlayOffset {
					triggered = true
				}
			}
			prevPlay = playPos
			if triggered {
				if d.onRowOutput != nil {
					d.onRowOutput(rowWave.Row)
				}
				break
			}
			time.Sleep(time.Millisecond * 1)
		}
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
