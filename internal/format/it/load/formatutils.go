package load

import (
	"gotracker/internal/format/it/layout"
	"gotracker/internal/format/it/playback"

	"github.com/gotracker/voice/pcm"
)

type readerFunc func(filename string, preferredSampleFormat ...pcm.SampleDataFormat) (*layout.Song, error)

func load(filename string, reader readerFunc, preferredSampleFormat ...pcm.SampleDataFormat) (*playback.Manager, error) {
	itSong, err := reader(filename, preferredSampleFormat...)
	if err != nil {
		return nil, err
	}

	m, err := playback.NewManager(itSong)

	return m, err
}
