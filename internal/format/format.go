package format

import (
	"errors"
	"os"

	"gotracker/internal/format/it"
	"gotracker/internal/format/mod"
	"gotracker/internal/format/s3m"
	"gotracker/internal/format/xm"
	"gotracker/internal/player/intf"

	"github.com/gotracker/voice/pcm"
)

var (
	supportedFormats = make(map[string]intf.Format)
)

// Load loads the a file into a playback manager
func Load(filename string, preferredSampleFormat ...pcm.SampleDataFormat) (intf.Playback, intf.Format, error) {
	for _, f := range supportedFormats {
		if playback, err := f.Load(filename, preferredSampleFormat...); err == nil {
			return playback, f, nil
		} else if os.IsNotExist(err) {
			return nil, nil, err
		}
	}
	return nil, nil, errors.New("unsupported format")
}

func init() {
	supportedFormats["s3m"] = s3m.S3M
	supportedFormats["mod"] = mod.MOD
	supportedFormats["xm"] = xm.XM
	supportedFormats["it"] = it.IT
}
