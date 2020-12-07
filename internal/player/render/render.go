package render

import "fmt"

type ChannelDisplay struct {
	Note       string
	Instrument string
	Volume     string
	Effect     string
}

type RowDisplay []ChannelDisplay

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

type RowRender struct {
	RenderData []byte
	Stop       bool
	Order      int
	Row        int
	RowText    RowDisplay
}
