package load

import (
	"gotracker/internal/player/intf"

	"github.com/gotracker/voice/pcm"
)

// IT loads an IT file
func IT(filename string, preferredSampleFormat ...pcm.SampleDataFormat) (intf.Playback, error) {
	return load(filename, readIT, preferredSampleFormat...)
}
