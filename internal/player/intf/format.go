package intf

import "gotracker/internal/format/settings"

// Format is an interface to a music file format loader
type Format interface {
	Load(filename string, s *settings.Settings) (Playback, error)
}
