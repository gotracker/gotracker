package load

import (
	"gotracker/internal/format/s3m/layout"
	"gotracker/internal/format/s3m/playback"
)

type readerFunc func(filename string) (*layout.Song, error)

func load(filename string, reader readerFunc) (*playback.Manager, error) {
	s3mSong, err := reader(filename)
	if err != nil {
		return nil, err
	}

	m, err := playback.NewManager(s3mSong)

	return m, err
}
