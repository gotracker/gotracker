package output

import (
	"gotracker/internal/player/intf"
	"gotracker/internal/player/render"

	"github.com/pkg/errors"
)

// RowOutputFunc defines the callback for when a row is output on the device
type RowOutputFunc func(row render.RowRender)

// Device is an interface to output device operations
type Device interface {
	Play(in <-chan render.RowRender)
	Close()
}

type deviceDetails struct {
	create         createOutputDeviceFunc
	kind           outputDeviceKind
	priority       outputDevicePriority
	featureDisable []intf.Feature
}

var (
	// DefaultOutputDeviceName is the default device name
	DefaultOutputDeviceName = "none"

	deviceMap = make(map[string]deviceDetails)
)

// CreateOutputDevice creates an output device based on the provided settings
func CreateOutputDevice(settings Settings) (Device, []intf.Feature, error) {
	if details, ok := deviceMap[settings.Name]; ok && details.create != nil {
		if dev, err := details.create(settings); err != nil {
			return nil, nil, err
		} else {
			return dev, details.featureDisable, nil
		}
	}

	return nil, nil, errors.New("device not supported")
}

type device struct {
	Device

	onRowOutput RowOutputFunc

	internal interface{}
}

// Settings is the settings for configuring an output device
type Settings struct {
	Name             string
	Channels         int
	SamplesPerSecond int
	BitsPerSample    int
	Filepath         string
	OnRowOutput      RowOutputFunc
}

func calculateOptimalDefaultOutputDeviceName() string {
	preferredPriority := outputDevicePriority(0)
	preferredName := "none"
	for name, details := range deviceMap {
		if details.priority > preferredPriority {
			preferredName = name
			preferredPriority = details.priority
		}
	}

	return preferredName
}

// Setup finalizes the output device preference system
func Setup() {
	DefaultOutputDeviceName = calculateOptimalDefaultOutputDeviceName()
}
