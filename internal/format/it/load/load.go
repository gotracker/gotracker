package load

import (
	"gotracker/internal/player/intf"
)

// IT loads an IT file
func IT(filename string) (intf.Playback, error) {
	return load(filename, readIT)
}
