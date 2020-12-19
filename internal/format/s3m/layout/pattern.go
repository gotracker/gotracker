package layout

import (
	s3mfile "github.com/heucuva/goaudiofile/music/tracked/s3m"

	"gotracker/internal/format/s3m/layout/channel"
	"gotracker/internal/player/intf"
)

// Pattern is the data for each pattern
type Pattern struct {
	intf.Pattern
	Packed s3mfile.PackedPattern
	Rows   [64]RowData
}

// GetRow returns the interface to the row at index `row`
func (p Pattern) GetRow(row uint8) intf.Row {
	return &p.Rows[row]
}

// GetRows returns the interfaces to all the rows in the pattern
func (p Pattern) GetRows() []intf.Row {
	rows := make([]intf.Row, len(p.Rows))
	for i, pr := range p.Rows {
		rows[i] = pr
	}
	return rows
}

// RowData is the data for each row
type RowData struct {
	intf.Row
	Channels [32]channel.Data
}

// GetChannels returns an interface to all the channels in the row
func (r RowData) GetChannels() []intf.ChannelData {
	c := make([]intf.ChannelData, len(r.Channels))
	for i := range r.Channels {
		c[i] = &r.Channels[i]
	}

	return c
}
