package player

import (
	"gotracker/internal/player/render"
	"gotracker/internal/player/state"
)

func RenderOneRow(ss *state.Song, sampler *render.Sampler) *render.RowRender {
	return ss.RenderOneRow(sampler)
}
