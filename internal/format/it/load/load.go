package load

import (
	"github.com/gotracker/gotracker/internal/format/settings"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// IT loads an IT file
func IT(filename string, s *settings.Settings) (intf.Playback, error) {
	return load(filename, readIT, s)
}
