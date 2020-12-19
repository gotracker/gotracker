// Package s3m does a thing.
package s3m

import (
	"gotracker/internal/format/s3m/load"
	"gotracker/internal/format/s3m/playback/util"
	"gotracker/internal/player/intf"
)

type format struct {
	intf.Format
}

var (
	// S3M is the exported interface to the S3M file loader
	S3M = format{}
)

// LoadMOD loads a MOD file and upgrades it into an S3M file internally
func LoadMOD(s intf.Song, filename string) error {
	return load.MOD(s, filename)
}

// GetBaseClockRate returns the base clock rate for the S3M player
func (f format) GetBaseClockRate() float32 {
	return util.S3MBaseClock
}

// Load loads an S3M file into the song state `s`
func (f format) Load(s intf.Song, filename string) error {
	return load.S3M(s, filename)
}
