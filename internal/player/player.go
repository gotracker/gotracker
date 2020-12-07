package player

import (
	"s3mplayer/internal/player/render"
	"s3mplayer/internal/player/state"
)

func RenderOneRow(ss *state.Song, sampler *render.Sampler) *render.RowRender {
	return ss.RenderOneRow(sampler)
}
