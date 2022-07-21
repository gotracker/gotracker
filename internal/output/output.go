package output

import (
	"errors"

	device "github.com/gotracker/gosound"

	playerFeature "github.com/gotracker/gotracker/internal/feature"
	"github.com/gotracker/playback/player/feature"
)

type devicePriority int

// the further down the list, the higher the priority
const (
	devicePriorityNone = devicePriority(iota)
	devicePriorityFile
	devicePriorityPulseAudio
	devicePriorityWinmm
	devicePriorityDirectSound
)

var (
	// DefaultOutputDeviceName is the default device name
	DefaultOutputDeviceName = "none"

	devicePriorityMap = make(map[string]devicePriority)
)

func calculateOptimalDefaultOutputDeviceName() string {
	preferredPriority := devicePriority(0)
	preferredName := "none"
	for name := range device.Map {
		if priority, ok := devicePriorityMap[name]; ok && priority > preferredPriority {
			preferredName = name
			preferredPriority = priority
		}
	}

	return preferredName
}

// CreateOutputDevice creates an output device based on the provided settings
func CreateOutputDevice(settings device.Settings) (device.Device, []feature.Feature, error) {
	d, err := device.CreateOutputDevice(settings)
	if err != nil {
		return nil, nil, err
	}

	if d == nil {
		return nil, nil, errors.New("could not create output device")
	}

	var featureDisable []feature.Feature

	kind := device.GetKind(d)
	switch kind {
	case device.KindFile:
		featureDisable = []feature.Feature{
			feature.SongLoop{Count: 0},
			playerFeature.PlayerSleepInterval{Enabled: false},
		}
	}

	return d, featureDisable, nil
}

// Setup finalizes the output device preference system
func Setup() {
	DefaultOutputDeviceName = calculateOptimalDefaultOutputDeviceName()
}

// DeviceInfo returns information about a device
type DeviceInfo struct {
	Priority int
	Kind     device.Kind
}

func GetOutputDevices() map[string]DeviceInfo {
	m := make(map[string]DeviceInfo)
	for k, v := range devicePriorityMap {
		if d, ok := device.Map[k]; ok {
			m[k] = DeviceInfo{
				Priority: int(v),
				Kind:     d.Kind,
			}
		}
	}
	return m
}

func init() {
	_ = devicePriorityNone // lint
	devicePriorityMap["file"] = devicePriorityFile
	devicePriorityMap["pulseaudio"] = devicePriorityPulseAudio
	devicePriorityMap["winmm"] = devicePriorityWinmm
	devicePriorityMap["directsound"] = devicePriorityDirectSound
}
