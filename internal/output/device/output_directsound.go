// +build windows,directsound

package device

import (
	"errors"
	"sync/atomic"
	"time"

	"gotracker/internal/audio/mixing"
	"gotracker/internal/output/device/win32"
	"gotracker/internal/output/device/win32/directsound"
)

type dsoundDevice struct {
	device

	ds           *directsound.DirectSound
	lpdsbPrimary *directsound.Buffer
	wfx          *win32.WAVEFORMATEX

	mix mixing.Mixer
}

func newDSoundDevice(settings Settings) (Device, error) {
	d := dsoundDevice{
		device: device{
			onRowOutput: settings.OnRowOutput,
		},
		mix: mixing.Mixer{
			Channels:      settings.Channels,
			BitsPerSample: settings.BitsPerSample,
		},
	}
	preferredDeviceName := ""

	ds, err := directsound.NewDSound(preferredDeviceName)
	if err != nil {
		return nil, err
	}
	d.ds = ds
	if d.ds == nil {
		return nil, errors.New("could not create directsound device")
	}

	lpdsbPrimary, wfx, err := ds.CreateSoundBufferPrimary(settings.Channels, settings.SamplesPerSecond, settings.BitsPerSample)
	if err != nil {
		ds.Close()
		return nil, err
	}
	d.lpdsbPrimary = lpdsbPrimary
	d.wfx = wfx

	return &d, nil
}

// Play starts the wave output device playing
func (d *dsoundDevice) Play(in <-chan *PremixData) {
	type RowWave struct {
		PlayOffset uint32
		Row        *PremixData
	}

	event, err := win32.CreateEvent()
	if err != nil {
		return
	}
	defer win32.CloseHandle(event)

	out := make(chan RowWave, 3)
	panmixer := mixing.GetPanMixer(d.mix.Channels)

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

	pn := []directsound.PositionNotify{
		{
			Offset:      uint32(playbackSize - int(d.wfx.NBlockAlign)),
			EventNotify: event,
		},
	}

	if err := notify.SetNotificationPositions(pn); err != nil {
		return
	}

	// play (looping)
	lpdsb.Play(true)

	done := make(chan struct{}, 1)
	defer close(done)

	go func() {
		writePos := 0
		for row := range in {
			var rowWave RowWave
			//_, writePos, err := lpdsb.GetCurrentPosition()
			numBytes := row.SamplesLen * int(d.wfx.NBlockAlign)
			segments, err := lpdsb.Lock(writePos%playbackSize, numBytes)
			if err != nil {
				continue
			}
			d.mix.FlattenTo(segments, panmixer, row.SamplesLen, row.Data)
			if err := lpdsb.Unlock(segments); err != nil {
				continue
			}
			rowWave.Row = row
			writePos += numBytes
			rowWave.PlayOffset = uint32(writePos)
			out <- rowWave
		}
		close(out)
	}()
	playBase := uint32(0)
	go func() {
		eventCh, closeFunc := win32.EventToChannel(event)
		defer closeFunc()
		for {
			select {
			case <-eventCh:
				atomic.AddUint32(&playBase, uint32(playbackSize))
			case <-done:
				return
			}
		}
	}()
	for rowWave := range out {
		for {
			playPos, _, _ := lpdsb.GetCurrentPosition()
			base := atomic.LoadUint32(&playBase)
			if playPos+base >= rowWave.PlayOffset {
				if d.onRowOutput != nil {
					d.onRowOutput(KindSoundCard, rowWave.Row)
				}
				break
			}
			time.Sleep(time.Millisecond * 1)
		}
	}
	done <- struct{}{}
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
	Map["directsound"] = deviceDetails{
		create: newDSoundDevice,
		kind:   KindSoundCard,
	}
}
