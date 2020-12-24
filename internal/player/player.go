package player

import (
	device "github.com/gotracker/gosound"

	"gotracker/internal/player/render"
	"gotracker/internal/player/state"
)

// RenderOneRow renders one row via the song state
func RenderOneRow(ss *state.Song, sampler *render.Sampler) (*device.PremixData, error) {
	return ss.RenderOneRow(sampler)
}
