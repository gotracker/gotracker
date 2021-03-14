package pattern

import (
	"gotracker/internal/index"
	"gotracker/internal/player/intf"
)

// RowData is the data for each row
type RowData struct {
	Channels []intf.ChannelData
}

// GetChannels returns an interface to all the channels in the row
func (r RowData) GetChannels() []intf.ChannelData {
	return r.Channels
}

// Rows is a list of row data (channels and whatnot)
type Rows []RowData

// GetRow returns the row at the specified row index from the list of rows
func (r Rows) GetRow(idx index.Row) intf.Row {
	return &r[int(idx)]
}

// NumRows returns the number of rows in this list of rows
func (r Rows) NumRows() int {
	return len(r)
}

// Pattern is the data for each pattern
type Pattern struct {
	Rows Rows
	Orig interface{}
}

// GetRow returns the interface to the row at index `row`
func (p Pattern) GetRow(row index.Row) intf.Row {
	return &p.Rows[row]
}

// GetRows returns the interfaces to all the rows in the pattern
func (p Pattern) GetRows() intf.Rows {
	return p.Rows
}
