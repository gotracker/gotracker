package load

import (
	formatutil "gotracker/internal/format/internal/util"
	"gotracker/internal/format/s3m/layout"
	"gotracker/internal/format/s3m/load/modconv"
	"gotracker/internal/player/intf"
)

func readMOD(filename string) (*layout.Song, error) {
	buffer, err := formatutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	f, err := modconv.Read(buffer)
	if err != nil {
		return nil, err
	}

	return convertS3MFileToSong(f)
}

// MOD loads a MOD file and upgrades it into an S3M file internally
func MOD(s intf.Song, filename string) error {
	return load(s, filename, readMOD)
}

// S3M loads an S3M file into the song state `s`
func S3M(s intf.Song, filename string) error {
	return load(s, filename, readS3M)
}
