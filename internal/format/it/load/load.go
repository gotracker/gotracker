package load

import (
	"gotracker/internal/format/settings"
	"gotracker/internal/player/intf"
)

// IT loads an IT file
func IT(filename string, s *settings.Settings) (intf.Playback, error) {
	return load(filename, readIT, s)
}
