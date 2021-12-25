// Package xm does a thing.
package xm

import (
	"gotracker/internal/format/settings"
	"gotracker/internal/format/xm/load"
	"gotracker/internal/player/intf"
)

type format struct{}

var (
	// XM is the exported interface to the XM file loader
	XM = format{}
)

// Load loads an XM file into a playback system
func (f format) Load(filename string, s *settings.Settings) (intf.Playback, error) {
	return load.XM(filename, s)
}
