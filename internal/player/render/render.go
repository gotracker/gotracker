package render

import (
	"fmt"
	"gotracker/internal/song"
	"strings"
)

// ChannelData is the data used by the ChannelFormatterFunc to render the source data from a tracker channel
type ChannelData any

// ChannelFormatterFunc takes the data from a channel and converts it to a string
type ChannelFormatterFunc func(song.ChannelData, bool) string

// RowDisplay is an array of ChannelDisplays
type RowDisplay struct {
	Channels   []song.ChannelData
	formatter  ChannelFormatterFunc
	longFormat bool
}

// NewRowText creates an array of ChannelDisplay information
func NewRowText(channels int, longFormat bool, channelFmtFunc ChannelFormatterFunc) RowDisplay {
	rd := RowDisplay{
		Channels:   make([]song.ChannelData, channels),
		formatter:  channelFmtFunc,
		longFormat: longFormat,
	}
	return rd
}

func (rt RowDisplay) String(options ...any) string {
	maxChannels := -1
	if len(options) > 0 {
		maxChannels = options[0].(int)
	}
	items := make([]string, 0, len(rt.Channels))
	for i, c := range rt.Channels {
		if maxChannels >= 0 && i >= maxChannels {
			break
		}
		items = append(items, fmt.Sprint(rt.formatter(c, rt.longFormat)))
	}
	return "|" + strings.Join(items, "|") + "|"
}

//RowRender is the final output of a single row's data
type RowRender struct {
	Order   int
	Row     int
	Tick    int
	RowText *RowDisplay
}
