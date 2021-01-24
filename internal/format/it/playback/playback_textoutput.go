package playback

import (
	"fmt"
	"strings"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/note"
	"gotracker/internal/player/render"
)

func xmChannelRender(cdata render.ChannelData) string {
	n := "..."
	i := ".."
	v := ".."
	e := "..."

	if data, ok := cdata.(*channel.Data); ok && data != nil {
		if data.HasNote() {
			nt := data.GetNote()
			switch nt {
			case note.ReleaseNote:
				n = "==="
			case note.StopNote:
				n = "^^^"
			case note.InvalidNote:
				n = "???"
			default:
				n = nt.String()
			}
		}

		if data.HasInstrument() {
			if inst := data.Instrument; inst != 0 {
				i = fmt.Sprintf("%0.2X", inst)
			}
		}

		if data.HasVolume() {
			vol := data.VolPan
			v = fmt.Sprintf("%0.2X", vol)
		}

		if data.HasCommand() {
			var c uint8
			switch {
			case data.Effect >= 0 && data.Effect <= 26:
				c = '@' + data.Effect
			default:
				panic("effect out of range")
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
	for ch, cs := range m.channels {
		if !m.song.IsChannelEnabled(ch) {
			continue
		}

		rowText.Channels[ch] = cs.TrackData
	}
	return &rowText
}
