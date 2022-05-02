package pattern

import (
	"gotracker/internal/song"
	"gotracker/internal/song/index"
)

// RowData is the data for each row
type RowData[TChannelData any] struct {
	Channels []TChannelData
}

// GetChannels returns an interface to all the channels in the row
func (r RowData[TChannelData]) GetChannels() []TChannelData {
	return r.Channels
}

// Rows is a list of row data (channels and whatnot)
type Rows[TChannelData any] []RowData[TChannelData]

// GetRow returns the row at the specified row index from the list of rows
func (r Rows[TChannelData]) GetRow(idx index.Row) song.Row[TChannelData] {
	return &r[int(idx)]
}

// NumRows returns the number of rows in this list of rows
func (r Rows[TChannelData]) NumRows() int {
	return len(r)
}

// Pattern is the data for each pattern
type Pattern[TChannelData any] struct {
	Rows Rows[TChannelData]
	Orig any
}

// GetRow returns the interface to the row at index `row`
func (p Pattern[TChannelData]) GetRow(row index.Row) song.Row[TChannelData] {
	return &p.Rows[row]
}

// GetRows returns the interfaces to all the rows in the pattern
func (p Pattern[TChannelData]) GetRows() song.Rows[TChannelData] {
	return p.Rows
}
