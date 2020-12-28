package mod

import (
	"gotracker/internal/format/s3m"
	"gotracker/internal/player/intf"
)

type format struct {
	intf.Format
}

var (
	// MOD is the exported interface to the MOD file loader
	MOD = format{}
)

// Load loads an MOD file into the song state `s`
func (f format) Load(filename string) (intf.Playback, error) {
	// we really just load the mod into an S3M layout, since S3M is essentially a superset
	return s3m.LoadMOD(filename)
}
