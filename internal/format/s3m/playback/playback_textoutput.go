package playback

import (
	"fmt"

	"github.com/gotracker/gomixing/volume"

	"gotracker/internal/player/render"
)

func (m *Manager) getRowText() render.RowDisplay {
	nCh := 0
	for ch := range m.channels {
		if !m.song.IsChannelEnabled(ch) {
			continue
		}
		nCh++
	}
	var rowText = render.NewRowText(nCh)
	for ch := range m.channels {
		if !m.song.IsChannelEnabled(ch) {
			continue
		}
		cs := &m.channels[ch]
		c := render.ChannelDisplay{
			Note:       cs.DisplayNote.String(),
			Instrument: "..",
			Volume:     "..",
			Effect:     "...",
		}

		if cs.DisplayInst != 0 {
			c.Instrument = fmt.Sprintf("%0.2d", cs.DisplayInst)
		}

		if cs.DisplayVolume != volume.VolumeUseInstVol {
			c.Volume = fmt.Sprintf("%0.2d", uint8(cs.DisplayVolume*64.0))
		}

		if cs.Cmd != nil {
			if cs.ActiveEffect != nil {
				c.Effect = cs.ActiveEffect.String()
			}
		}
		rowText[ch] = c
	}
	return rowText
}
