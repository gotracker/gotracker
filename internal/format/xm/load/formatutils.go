package load

import (
	"gotracker/internal/format/settings"
	"gotracker/internal/format/xm/layout"
	"gotracker/internal/format/xm/playback"
)

type readerFunc func(filename string, s *settings.Settings) (*layout.Song, error)

func load(filename string, reader readerFunc, s *settings.Settings) (*playback.Manager, error) {
	xmSong, err := reader(filename, s)
	if err != nil {
		return nil, err
	}

	m, err := playback.NewManager(xmSong)

	return m, err
}
