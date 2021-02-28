package load

import (
	"gotracker/internal/format/it/layout"
	"gotracker/internal/format/it/playback"
)

type readerFunc func(filename string) (*layout.Song, error)

func load(filename string, reader readerFunc) (*playback.Manager, error) {
	itSong, err := reader(filename)
	if err != nil {
		return nil, err
	}

	m, err := playback.NewManager(itSong)

	return m, err
}
