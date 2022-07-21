package device

import (
	"context"
	"errors"
	"path"
	"strings"

	deviceCommon "github.com/gotracker/gotracker/internal/output/device/common"
	deviceFile "github.com/gotracker/gotracker/internal/output/device/file"
	"github.com/gotracker/playback/output"
)

const fileName = "file"

type fileDevice struct {
	device
	processor deviceFile.File
}

func (fileDevice) GetKind() deviceCommon.Kind {
	return deviceCommon.KindFile
}

// Name returns the device name
func (fileDevice) Name() string {
	return fileName
}

func (d *fileDevice) Play(in <-chan *output.PremixData) error {
	var onWrittenCallback deviceFile.WrittenCallback
	if d.onRowOutput != nil {
		onWrittenCallback = func(data *output.PremixData) {
			d.onRowOutput(deviceCommon.KindFile, data)
		}
	}
	return d.processor.Play(in, onWrittenCallback)
}

func (d *fileDevice) PlayWithCtx(ctx context.Context, in <-chan *output.PremixData) error {
	var onWrittenCallback deviceFile.WrittenCallback
	if d.onRowOutput != nil {
		onWrittenCallback = func(data *output.PremixData) {
			d.onRowOutput(deviceCommon.KindFile, data)
		}
	}
	return d.processor.PlayWithCtx(ctx, in, onWrittenCallback)
}

func (d *fileDevice) Close() error {
	return d.processor.Close()
}

func newFileDevice(settings deviceCommon.Settings) (Device, error) {
	ext := strings.ToLower(path.Ext(settings.Filepath))
	if factory, ok := deviceFile.GetFileDevice(ext); ok && factory != nil {
		processor, err := factory(settings)
		if err != nil {
			return nil, err
		}
		dev := fileDevice{
			device: device{
				onRowOutput: settings.OnRowOutput,
			},
			processor: processor,
		}
		return &dev, nil
	}

	return nil, errors.New("unsupported output format")
}

func init() {
	Map[fileName] = deviceDetails{
		create: newFileDevice,
		Kind:   deviceCommon.KindFile,
	}
}
