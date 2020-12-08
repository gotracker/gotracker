// +build linux

package output

import (
	"strings"
)

const (
	// DefaultOutputDeviceName specifies the default device for linux playback
	DefaultOutputDeviceName = "file"
)

// CreateOutputDevice creates an output device based on the provided settings
func CreateOutputDevice(settings Settings) (Device, error) {
	switch strings.ToLower(settings.Name) {
	}
	return createGeneralDevice(settings)
}
