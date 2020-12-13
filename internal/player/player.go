package player

import (
	"gotracker/internal/player/render"
	"gotracker/internal/player/state"
)

// RenderOneRow renders one row via the song state
func RenderOneRow(ss *state.Song, sampler *render.Sampler) *render.RowRender {
	return ss.RenderOneRow(sampler)
}
