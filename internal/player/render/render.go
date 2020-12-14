package render

import (
	"fmt"
	"gotracker/internal/player/panning"
	"gotracker/internal/player/render/mixer"
	"gotracker/internal/player/volume"
)

// ChannelDisplay is a render output of tracker channel information
type ChannelDisplay struct {
	Note       string
	Instrument string
	Volume     string
	Effect     string
}

// RowDisplay is an array of ChannelDisplays
type RowDisplay []ChannelDisplay

// NewRowText creates an array of ChannelDisplay information
func NewRowText(channels int) RowDisplay {
	return make([]ChannelDisplay, channels)
}

func (rt RowDisplay) String(options ...interface{}) string {
	maxChannels := -1
	if len(options) > 0 {
		maxChannels = options[0].(int)
	}
	var str string
	for i, c := range rt {
		if maxChannels >= 0 && i >= maxChannels {
			break
		}
		str += fmt.Sprintf("|%s %s %s %s", c.Note, c.Instrument, c.Volume, c.Effect)
	}
	return str + "|"
}

// Data is a single buffer of data at a specific panning position
type Data struct {
	Data       mixer.MixBuffer
	Pan        panning.Position
	Volume     volume.Volume
	SamplesLen int
	Flush      func()
}

// ChannelData is a single channel's data
type ChannelData []Data

//RowRender is the final output of a single row's data
type RowRender struct {
	RenderData []ChannelData
	SamplesLen int
	Stop       bool
	Order      int
	Row        int
	RowText    RowDisplay
}
