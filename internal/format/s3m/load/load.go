package load

import (
	formatutil "github.com/gotracker/gotracker/internal/format/internal/util"
	"github.com/gotracker/gotracker/internal/format/s3m/layout"
	"github.com/gotracker/gotracker/internal/format/s3m/load/modconv"
	"github.com/gotracker/gotracker/internal/format/settings"
	"github.com/gotracker/gotracker/internal/player/intf"
)

func readMOD(filename string, s *settings.Settings) (*layout.Song, error) {
	buffer, err := formatutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	f, err := modconv.Read(buffer)
	if err != nil {
		return nil, err
	}

	return convertS3MFileToSong(f, func(patNum int) uint8 {
		return 64
	}, s)
}

// MOD loads a MOD file and upgrades it into an S3M file internally
func MOD(filename string, s *settings.Settings) (intf.Playback, error) {
	return load(filename, readMOD, s)
}

// S3M loads an S3M file into a new Playback object
func S3M(filename string, s *settings.Settings) (intf.Playback, error) {
	return load(filename, readS3M, s)
}
