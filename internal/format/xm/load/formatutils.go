package load

import (
	"gotracker/internal/format/xm/layout"
	"gotracker/internal/format/xm/playback"
)

type readerFunc func(filename string) (*layout.Song, error)

func load(filename string, reader readerFunc) (*playback.Manager, error) {
	xmSong, err := reader(filename)
	if err != nil {
		return nil, err
	}

	m := playback.NewManager(xmSong)

	return m, nil
}
