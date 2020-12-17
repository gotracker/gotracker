package player

import (
	"gotracker/internal/module/player/render"
	"gotracker/internal/module/player/state"
	"gotracker/internal/output/device"
)

// RenderOneRow renders one row via the song state
func RenderOneRow(ss *state.Song, sampler *render.Sampler) (*device.PremixData, error) {
	return ss.RenderOneRow(sampler)
}
