package playback

import (
	"fmt"
	"strings"

	"github.com/gotracker/goaudiofile/music/tracked/s3m"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/render"
)

func s3mChannelRender(cdata render.ChannelData) string {
	n := "..."
	i := ".."
	v := ".."
	e := "..."

	if data, ok := cdata.(*channel.Data); ok && data != nil {
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

		if data.HasCommand() {
			e = fmt.Sprintf("%c%0.2X", '@'+data.Command, data.Info)
		}
	}

	return strings.Join([]string{n, i, v, e}, " ")
}

func (m *Manager) getRowText() render.RowDisplay {
	nCh := 0
	for ch := range m.channels {
		if !m.song.IsChannelEnabled(ch) {
			continue
		}
		nCh++
	}
	var rowText = render.NewRowText(nCh, s3mChannelRender)
	for ch := range m.channels {
		if !m.song.IsChannelEnabled(ch) {
			continue
		}
		cs := &m.channels[ch]

		rowText.Channels[ch] = cs.Cmd
	}
	return rowText
}
