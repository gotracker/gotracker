package mod

import (
	"gotracker/internal/format/s3m"
	"gotracker/internal/format/settings"
	"gotracker/internal/player/intf"
)

type format struct{}

var (
	// MOD is the exported interface to the MOD file loader
	MOD = format{}
)

// Load loads an MOD file into the song state
func (f format) Load(filename string, s *settings.Settings) (intf.Playback, error) {
	// we really just load the mod into an S3M layout, since S3M is essentially a superset
	return s3m.LoadMOD(filename, s)
}
