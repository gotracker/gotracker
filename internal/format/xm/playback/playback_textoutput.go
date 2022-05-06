package playback

import (
	"fmt"
	"strings"

	"github.com/gotracker/gotracker/internal/format/xm/layout/channel"
	"github.com/gotracker/gotracker/internal/player/render"
	"github.com/gotracker/gotracker/internal/song"
	"github.com/gotracker/gotracker/internal/song/note"
)

func xmChannelRender(cdata song.ChannelData, longChannelOutput bool) string {
	n := "..."
	i := "  "
	v := ".."
	e := "..."

	if data, _ := cdata.(*channel.Data); data != nil {
		if data.HasNote() {
			nt := data.GetNote()
			switch note.Type(nt) {
			case note.SpecialTypeRelease:
				n = "== "
			case note.SpecialTypeNormal:
				n = nt.String()
			default:
				n = "???"
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

		if data.HasCommand() {
			var c uint8
			switch {
			case data.Effect <= 9:
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
	rowText := render.NewRowText(nCh, true, xmChannelRender)
	for ch, cs := range m.channels {
		if !m.song.IsChannelEnabled(ch) {
			continue
		}

		rowText.Channels[ch] = cs.TrackData
	}
	return &rowText
}
