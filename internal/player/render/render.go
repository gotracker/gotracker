package render

import (
	"strings"

	"github.com/gotracker/gotracker/internal/song"
)

// RowDisplay is an array of ChannelDisplays
type RowDisplay[TChannelData song.ChannelData] struct {
	Channels   []TChannelData
	longFormat bool
}

// NewRowText creates an array of ChannelDisplay information
func NewRowText[TChannelData song.ChannelData](channels int, longFormat bool) RowDisplay[TChannelData] {
	rd := RowDisplay[TChannelData]{
		Channels:   make([]TChannelData, channels),
		longFormat: longFormat,
	}
	return rd
}

func (rt RowDisplay[TChannelData]) String(options ...any) string {
	maxChannels := -1
	if len(options) > 0 {
		maxChannels = options[0].(int)
	}
	items := make([]string, 0, len(rt.Channels))
	for i, c := range rt.Channels {
		if maxChannels >= 0 && i >= maxChannels {
			break
		}
		if rt.longFormat {
			items = append(items, c.String())
		} else {
			items = append(items, c.ShortString())
		}
	}
	return "|" + strings.Join(items, "|") + "|"
}

type RowStringer interface {
	String(options ...any) string
}

//RowRender is the final output of a single row's data
type RowRender struct {
	Order   int
	Row     int
	Tick    int
	RowText RowStringer
}
