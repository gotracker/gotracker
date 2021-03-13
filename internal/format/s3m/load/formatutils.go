package load

import (
	"gotracker/internal/format/s3m/layout"
	"gotracker/internal/format/s3m/playback"

	"github.com/gotracker/voice/pcm"
)

type readerFunc func(filename string, preferredSampleFormat ...pcm.SampleDataFormat) (*layout.Song, error)

func load(filename string, reader readerFunc, preferredSampleFormat ...pcm.SampleDataFormat) (*playback.Manager, error) {
	s3mSong, err := reader(filename, preferredSampleFormat...)
	if err != nil {
		return nil, err
	}

	m, err := playback.NewManager(s3mSong)

	return m, err
}
