// Package it does a thing.
package it

import (
	"gotracker/internal/format/it/load"
	"gotracker/internal/player/intf"

	"github.com/gotracker/voice/pcm"
)

type format struct {
	intf.Format
}

var (
	// IT is the exported interface to the IT file loader
	IT = format{}
)

// Load loads an IT file into a playback system
func (f format) Load(filename string, preferredSampleFormat ...pcm.SampleDataFormat) (intf.Playback, error) {
	return load.IT(filename, preferredSampleFormat...)
}
