package device

import (
	"errors"
	"path"
	"strings"
)

var (
	fileDeviceMap = make(map[string]createOutputDeviceFunc)
)

func newFileDevice(settings Settings) (Device, error) {
	ext := strings.ToLower(path.Ext(settings.Filepath))
	if create, ok := fileDeviceMap[ext]; ok && create != nil {
		return create(settings)
	}

	return nil, errors.New("unsupported output format")
}

func init() {
	Map["file"] = deviceDetails{
		create: newFileDevice,
		kind:   KindFile,
	}
}
