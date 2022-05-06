package load

import (
	"github.com/gotracker/gotracker/internal/format/s3m/layout"
	"github.com/gotracker/gotracker/internal/format/s3m/playback"
	"github.com/gotracker/gotracker/internal/format/settings"
)

type readerFunc func(filename string, s *settings.Settings) (*layout.Song, error)

func load(filename string, reader readerFunc, s *settings.Settings) (*playback.Manager, error) {
	s3mSong, err := reader(filename, s)
	if err != nil {
		return nil, err
	}

	m, err := playback.NewManager(s3mSong)

	return m, err
}
