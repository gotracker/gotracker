package load

import (
	"github.com/gotracker/gotracker/internal/format/settings"
	"github.com/gotracker/gotracker/internal/player/intf"
)

// XM loads an XM file and upgrades it into an XM file internally
func XM(filename string, s *settings.Settings) (intf.Playback, error) {
	return load(filename, readXM, s)
}
