package format

import (
	"errors"
	"os"

	"gotracker/internal/format/it"
	"gotracker/internal/format/mod"
	"gotracker/internal/format/s3m"
	"gotracker/internal/format/settings"
	"gotracker/internal/format/xm"
	"gotracker/internal/player/intf"
	"gotracker/internal/song"
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
