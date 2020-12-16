// +build flac

package output

import (
	"bufio"
	"gotracker/internal/player/render"
	"gotracker/internal/player/render/mixer"
	"os"

	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/frame"
	"github.com/mewkiz/flac/meta"
	"github.com/pkg/errors"
)

type fileDeviceFlac struct {
	device
	mix              mixer.Mixer
	samplesPerSecond int

	f *os.File
	w *bufio.Writer
}

func newFileFlacDevice(settings Settings) (Device, error) {
	fd := fileDeviceFlac{
		device: device{
			onRowOutput: settings.OnRowOutput,
		},
		mix: mixer.Mixer{
			Channels:      settings.Channels,
			BitsPerSample: settings.BitsPerSample,
		},
		samplesPerSecond: settings.SamplesPerSecond,
	}
	f, err := os.OpenFile(settings.Filepath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	if f == nil {
		return nil, errors.New("unexpected file error")
	}

	fd.f = f

	return &fd, nil
}

// Play starts the wave output device playing
func (d *fileDeviceFlac) Play(in <-chan render.RowRender) {
	w := bufio.NewWriter(d.f)
	d.w = w
	// Encode FLAC stream.
	si := &meta.StreamInfo{
		BlockSizeMin:  16,
		BlockSizeMax:  65535,
		SampleRate:    uint32(d.samplesPerSecond),
		NChannels:     uint8(d.mix.Channels),
		BitsPerSample: uint8(d.mix.BitsPerSample),
	}
	enc, err := flac.NewEncoder(w, si)
	if err != nil {
		return
	}
	defer enc.Close()

	panmixer := mixer.GetPanMixer(d.mix.Channels)
	var channels frame.Channels
	switch d.mix.Channels {
	case 1:
		channels = frame.ChannelsMono
	case 2:
		channels = frame.ChannelsLR
	case 4:
		channels = frame.ChannelsLRLsRs
	}

	for row := range in {
		mixedData := d.mix.FlattenToInts(panmixer, row.SamplesLen, row.RenderData)
		subframes := make([]*frame.Subframe, d.mix.Channels)
		for i := range subframes {
			subframe := &frame.Subframe{
				SubHeader: frame.SubHeader{
					Pred: frame.PredVerbatim,
				},
				Samples:  mixedData[i],
				NSamples: row.SamplesLen,
			}
			subframes[i] = subframe
		}
		for _, subframe := range subframes {
			sample := subframe.Samples[0]
			constant := true
			for _, s := range subframe.Samples[1:] {
				if sample != s {
					constant = false
				}
			}
			if constant {
				subframe.SubHeader.Pred = frame.PredConstant
			}
		}

		fr := &frame.Frame{
			Header: frame.Header{
				HasFixedBlockSize: false,
				BlockSize:         uint16(row.SamplesLen),
				SampleRate:        uint32(d.samplesPerSecond),
				Channels:          channels,
				BitsPerSample:     uint8(d.mix.BitsPerSample),
			},
			Subframes: subframes,
		}
		if err := enc.WriteFrame(fr); err != nil {
			panic(err)
		}
		if d.onRowOutput != nil {
			d.onRowOutput(DeviceKindFile, row)
		}
	}
}

// Close closes the wave output device
func (d *fileDeviceFlac) Close() {
	d.w.Flush()
	d.w = nil
	d.f.Close()
}

func init() {
	fileDeviceMap[".flac"] = newFileFlacDevice
}
