package load

import (
	"gotracker/internal/player/intf"

	"github.com/gotracker/voice/pcm"
)

// XM loads an XM file and upgrades it into an XM file internally
func XM(filename string, preferredSampleFormat ...pcm.SampleDataFormat) (intf.Playback, error) {
	return load(filename, readXM, preferredSampleFormat...)
}
