package format

import (
	"errors"
	"os"

	"github.com/gotracker/gotracker/internal/format/it"
	"github.com/gotracker/gotracker/internal/format/mod"
	"github.com/gotracker/gotracker/internal/format/s3m"
	"github.com/gotracker/gotracker/internal/format/settings"
	"github.com/gotracker/gotracker/internal/format/xm"
	"github.com/gotracker/gotracker/internal/player/intf"
	"github.com/gotracker/gotracker/internal/song"
)

var (
	supportedFormats = make(map[string]intf.Format[song.ChannelData])
)

// Load loads the a file into a playback manager
func Load(filename string, options ...settings.OptionFunc) (intf.Playback, intf.Format[song.ChannelData], error) {
	s := &settings.Settings{}
	for _, opt := range options {
		if err := opt(s); err != nil {
			return nil, nil, err
		}
	}

	for _, f := range supportedFormats {
		if playback, err := f.Load(filename, s); err == nil {
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
