package format

import (
	"errors"
	"os"

	"gotracker/internal/format/mod"
	"gotracker/internal/format/s3m"
	"gotracker/internal/player/intf"
)

var (
	supportedFormats = make(map[string]intf.Format)
)

// Load loads the a file into a playback manager
func Load(filename string) (intf.Playback, intf.Format, error) {
	for _, fmt := range supportedFormats {
		if playback, err := fmt.Load(filename); err == nil {
			return playback, fmt, nil
		} else if os.IsNotExist(err) {
			return nil, nil, err
		}
	}
	return nil, nil, errors.New("unsupported format")
}

func init() {
	supportedFormats["s3m"] = s3m.S3M
	supportedFormats["mod"] = mod.MOD
}
