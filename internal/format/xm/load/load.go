package load

import (
	"gotracker/internal/player/intf"
)

// XM loads an XM file and upgrades it into an XM file internally
func XM(filename string) (intf.Playback, error) {
	return load(filename, readXM)
}
