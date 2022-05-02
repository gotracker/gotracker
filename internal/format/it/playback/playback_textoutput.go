package playback

import (
	"fmt"
	"strings"

	"gotracker/internal/format/it/layout/channel"
	"gotracker/internal/player/render"
	"gotracker/internal/song"
	"gotracker/internal/song/note"
)

func itChannelRender(cdata song.ChannelData, longChannelOutput bool) string {
	n := "..."
	i := ".."
	v := ".."
	e := "..."

	if data, _ := cdata.(*channel.Data); data != nil {
		if data.HasNote() {
			nt := data.GetNote()
			switch note.Type(nt) {
			case note.SpecialTypeRelease:
				n = "==="
			case note.SpecialTypeStop:
				n = "^^^"
			case note.SpecialTypeNormal:
				n = nt.String()
			default:
				n = "???"
			}
		}

		if longChannelOutput {
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
				case data.Effect <= 26:
					c = '@' + data.Effect
				default:
					panic("effect out of range")
				}
				e = fmt.Sprintf("%c%0.2X", c, data.EffectParameter)
			}
		}
	}

	if longChannelOutput {
		return strings.Join([]string{n, i, v, e}, " ")

	}

	return n
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

		rowText.Channels[ch] = cs.TrackData
	}
	return &rowText
}
