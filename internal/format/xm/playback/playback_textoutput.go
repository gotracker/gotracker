package playback

import (
	"fmt"
	"strings"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/note"
	"gotracker/internal/player/render"
)

func xmChannelRender(cdata render.ChannelData) string {
	n := "..."
	i := "  "
	v := ".."
	e := "..."

	if data, ok := cdata.(*channel.Data); ok && data != nil {
		if data.HasNote() {
			nt := data.GetNote()
			if nt != note.StopNote {
				n = nt.String()
			} else {
				n = "== "
			}
		}

		if data.HasInstrument() {
			if inst := data.Instrument; inst != 0 {
				i = fmt.Sprintf("%X", inst)
				for len(i) < 2 {
					i = " " + i
				}
			}
		}

		if data.HasVolume() {
			if vol := data.Volume; vol != 255 {
				v = fmt.Sprintf("%0.2X", vol)
			}
		}

		if data.HasEffect() {
			var c uint8
			switch {
			case data.Effect >= 0 && data.Effect <= 9:
				c = '0' + data.Effect
			case data.Effect >= 10 && data.Effect < 36:
				c = 'A' + (data.Effect - 10)
			}
			e = fmt.Sprintf("%c%0.2X", c, data.EffectParameter)
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
	rowText := render.NewRowText(nCh, xmChannelRender)
	for ch := range m.channels {
		if !m.song.IsChannelEnabled(ch) {
			continue
		}
		cs := &m.channels[ch]

		rowText.Channels[ch] = cs.Cmd
	}
	return &rowText
}
