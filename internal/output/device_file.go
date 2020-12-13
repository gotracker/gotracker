package output

import (
	"gotracker/internal/player/intf"
	"path"
	"strings"

	"github.com/pkg/errors"
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
	deviceMap["file"] = deviceDetails{
		create:   newFileDevice,
		kind:     outputDeviceKindFile,
		priority: outputDevicePriorityFile,
		featureDisable: []intf.Feature{
			intf.FeaturePatternLoop,
		},
	}
}
