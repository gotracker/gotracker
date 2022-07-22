package file

import (
	"context"

	deviceCommon "github.com/gotracker/gotracker/internal/output/device/common"
	"github.com/gotracker/playback/output"
)

var (
	fileDeviceMap = make(map[string]FileFactory)
)

type FileFactory func(settings deviceCommon.Settings) (File, error)

type WrittenCallback func(data *output.PremixData)

type File interface {
	Play(in <-chan *output.PremixData, onWrittenCallback WrittenCallback) error
	PlayWithCtx(ctx context.Context, in <-chan *output.PremixData, onWrittenCallback WrittenCallback) error
	Close() error
}

func GetFileDevice(extension string) (FileFactory, bool) {
	factory, ok := fileDeviceMap[extension]
	return factory, ok
}
