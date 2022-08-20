package file

import (
	"bufio"
	"context"
	"errors"
	"os"

	"github.com/gotracker/gomixing/mixing"
	"github.com/gotracker/gomixing/sampling"
	deviceCommon "github.com/gotracker/gotracker/internal/output/device/common"
	"github.com/gotracker/playback/output"
)

type filePCM struct {
	mix     mixing.Mixer
	sampFmt sampling.Format

	f  *os.File
	w  *bufio.Writer
	sz uint32
}

func newFilePCMDevice(settings deviceCommon.Settings) (File, error) {
	fd := filePCM{
		mix: mixing.Mixer{
			Channels: settings.Channels,
		},
	}
	switch settings.BitsPerSample {
	case 8:
		fd.sampFmt = sampling.Format8BitUnsigned
	case 16:
		fd.sampFmt = sampling.Format16BitLESigned
	}

	f, err := os.OpenFile(settings.Filepath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	if f == nil {
		return nil, errors.New("unexpected file error")
	}

	w := bufio.NewWriter(f)

	fd.f = f
	fd.w = w

	return &fd, nil
}

// Play starts the wave output device playing
func (d *filePCM) Play(in <-chan *output.PremixData, onWrittenCallback WrittenCallback) error {
	return d.PlayWithCtx(context.Background(), in, onWrittenCallback)
}

// PlayWithCtx starts the wave output device playing
func (d *filePCM) PlayWithCtx(ctx context.Context, in <-chan *output.PremixData, onWrittenCallback WrittenCallback) error {
	panmixer := mixing.GetPanMixer(d.mix.Channels)
	if panmixer == nil {
		return errors.New("invalid pan mixer - check channel count")
	}

	myCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		select {
		case <-myCtx.Done():
			return myCtx.Err()
		case row, ok := <-in:
			if !ok {
				return nil
			}
			mixedData := d.mix.Flatten(panmixer, row.SamplesLen, row.Data, row.MixerVolume, d.sampFmt)
			sz, err := d.w.Write(mixedData)
			if err != nil {
				return err
			}
			d.sz += uint32(sz)
			if onWrittenCallback != nil {
				onWrittenCallback(row)
			}
		}
	}
}

// Close closes the wave output device
func (d *filePCM) Close() error {
	d.w.Flush()
	d.w = nil
	return d.f.Close()
}

func init() {
	fileDeviceMap[".pcm"] = newFilePCMDevice
	fileDeviceMap[".raw"] = newFilePCMDevice
}
