package intf

import "github.com/gotracker/gotracker/internal/format/settings"

// Format is an interface to a music file format loader
type Format[TChannelData any] interface {
	Load(filename string, s *settings.Settings) (Playback, error)
}
