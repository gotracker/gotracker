package layout

import (
	xmfile "github.com/gotracker/goaudiofile/music/tracked/xm"

	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// RowData is the data for each row
type RowData struct {
	intf.Row
	Channels []channel.Data
}

// GetChannels returns an interface to all the channels in the row
func (r RowData) GetChannels() []intf.ChannelData {
	c := make([]intf.ChannelData, len(r.Channels))
	for i := range r.Channels {
		c[i] = &r.Channels[i]
	}

	return c
}

// Rows is a list of row data (channels and whatnot)
type Rows []RowData

// GetRow returns the row at the specified row index from the list of rows
func (r Rows) GetRow(idx intf.RowIdx) intf.Row {
	return &r[int(idx)]
}

// NumRows returns the number of rows in this list of rows
func (r Rows) NumRows() int {
	return len(r)
}

// Pattern is the data for each pattern
type Pattern struct {
	intf.Pattern
	Orig xmfile.Pattern
	Rows Rows
}

// GetRow returns the interface to the row at index `row`
func (p Pattern) GetRow(row intf.RowIdx) intf.Row {
	return &p.Rows[row]
}

// GetRows returns the interfaces to all the rows in the pattern
func (p Pattern) GetRows() intf.Rows {
	return p.Rows
}
