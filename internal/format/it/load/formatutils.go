package load

import (
	"github.com/gotracker/gotracker/internal/format/it/layout"
	"github.com/gotracker/gotracker/internal/format/it/playback"
	"github.com/gotracker/gotracker/internal/format/settings"
)

type readerFunc func(filename string, s *settings.Settings) (*layout.Song, error)

func load(filename string, reader readerFunc, s *settings.Settings) (*playback.Manager, error) {
	itSong, err := reader(filename, s)
	if err != nil {
		return nil, err
	}

	m, err := playback.NewManager(itSong)

	return m, err
}
