package load

import (
	"gotracker/internal/format/xm/layout"
	"gotracker/internal/format/xm/playback"

	"github.com/gotracker/voice/pcm"
)

type readerFunc func(filename string, preferredSampleFormat ...pcm.SampleDataFormat) (*layout.Song, error)

func load(filename string, reader readerFunc, preferredSampleFormat ...pcm.SampleDataFormat) (*playback.Manager, error) {
	xmSong, err := reader(filename, preferredSampleFormat...)
	if err != nil {
		return nil, err
	}

	m, err := playback.NewManager(xmSong)

	return m, err
}
