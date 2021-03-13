// Package xm does a thing.
package xm

import (
	"gotracker/internal/format/xm/load"
	"gotracker/internal/player/intf"

	"github.com/gotracker/voice/pcm"
)

type format struct {
	intf.Format
}

var (
	// XM is the exported interface to the XM file loader
	XM = format{}
)

// Load loads an XM file into a playback system
func (f format) Load(filename string, preferredSampleFormat ...pcm.SampleDataFormat) (intf.Playback, error) {
	return load.XM(filename, preferredSampleFormat...)
}
