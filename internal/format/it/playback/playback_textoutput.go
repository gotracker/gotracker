package playback

import (
	"github.com/gotracker/gotracker/internal/format/it/layout/channel"
	"github.com/gotracker/gotracker/internal/player/render"
	"github.com/gotracker/gotracker/internal/song"
)

func itChannelRender(cdata song.ChannelData, longChannelOutput bool) string {
	data, _ := cdata.(*channel.Data)
	return channel.DataToString(data, longChannelOutput)
}

func (m *Manager) getRowText() *render.RowDisplay {
	nCh := 0
	for ch := range m.channels {
		if !m.song.IsChannelEnabled(ch) {
			continue
		}
		nCh++
	}
	rowText := render.NewRowText(nCh, m.longChannelOutput, itChannelRender)
	for ch, cs := range m.channels {
		if !m.song.IsChannelEnabled(ch) {
			continue
		}

		rowText.Channels[ch] = cs.GetData()
	}
	return &rowText
}
