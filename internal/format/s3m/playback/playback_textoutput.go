package playback

import (
	"fmt"
	"strings"

	"github.com/gotracker/goaudiofile/music/tracked/s3m"

	"github.com/gotracker/gotracker/internal/format/s3m/layout/channel"
	"github.com/gotracker/gotracker/internal/player/render"
	"github.com/gotracker/gotracker/internal/song"
)

func s3mChannelRender(cdata song.ChannelData, longChannelOutput bool) string {
	n := "..."
	i := ".."
	v := ".."
	e := "..."

	if data, _ := cdata.(*channel.Data); data != nil {
		if data.HasNote() {
			n = data.GetNote().String()
		}

		if data.HasInstrument() {
			if inst := data.Instrument; inst != 0 {
				i = fmt.Sprintf("%0.2d", inst)
			}
		}

		if data.HasVolume() {
			if vol := data.Volume; vol != s3m.EmptyVolume {
				v = fmt.Sprintf("%0.2d", vol)
			}
		}

		if data.HasCommand() && data.Command != 0 {
			e = fmt.Sprintf("%c%0.2X", '@'+data.Command, data.Info)
		}
	}

	return strings.Join([]string{n, i, v, e}, " ")
}

func (m *Manager) getRowText() *render.RowDisplay {
	nCh := 0
	for ch := range m.channels {
		if !m.song.IsChannelEnabled(ch) {
			continue
		}
		nCh++
	}
	rowText := render.NewRowText(nCh, true, s3mChannelRender)
	for ch, cs := range m.channels {
		if !m.song.IsChannelEnabled(ch) {
			continue
		}

		rowText.Channels[ch] = cs.TrackData
	}
	return &rowText
}
