package state

import (
	"gotracker/internal/player/intf"
)

// PatternNum is an order pattern number
type PatternNum uint8

const (
	// NextPattern allows the order system the ability to kick to the next pattern
	NextPattern = PatternNum(254)
	// InvalidPattern specifies an invalid pattern
	InvalidPattern = PatternNum(255)
)

// RowSettings is the settings for the current pattern state
type RowSettings struct {
	Ticks int
	Tempo int
}

// Row is a specification of the current row data
type Row struct {
	intf.Row
	Channels [32]intf.ChannelData
}

// PatternState is the current pattern state
type PatternState struct {
	CurrentOrder uint8
	CurrentRow   uint8

	Row RowSettings

	RowHasPatternDelay bool
	PatternDelay       int
	FinePatternDelay   int

	Patterns intf.Patterns
	Orders   []uint8

	LoopStart   uint8
	LoopEnd     uint8
	LoopTotal   uint8
	LoopEnabled bool
	LoopCount   uint8
}

// GetPatNum returns the current pattern number
func (state *PatternState) GetPatNum() PatternNum {
	if int(state.CurrentOrder) > len(state.Orders) {
		return InvalidPattern
	}
	return PatternNum(state.Orders[state.CurrentOrder])
}

// GetNumRows returns the number of rows in the current pattern
func (state *PatternState) GetNumRows() uint8 {
	rows := state.GetRows()
	return uint8(len(rows))
}

// WantsStop returns true when the current pattern wants to end the song
func (state *PatternState) WantsStop() bool {
	if state.GetPatNum() == InvalidPattern {
		return true
	}
	return false
}

// NextOrder travels to the next pattern in the order list
func (state *PatternState) NextOrder() {
	state.CurrentOrder++
	state.CurrentRow = 0
}

// NextRow travels to the next row in the pattern
// or the next order in the order list if the last row has been exhausted
func (state *PatternState) NextRow() {
	var patNum = state.GetPatNum()
	if patNum == InvalidPattern {
		return
	}

	if patNum == NextPattern {
		state.NextOrder()
		return
	}

	state.CurrentRow++
	if state.CurrentRow >= state.GetNumRows() {
		state.NextOrder()
		return
	}
}

// GetRow returns the current row
func (state *PatternState) GetRow() *Row {
	var patNum = state.GetPatNum()
	switch patNum {
	case InvalidPattern:
		return nil
	case NextPattern:
		{
			state.NextRow()
			return state.GetRow()
		}
	default:
		{
			var pattern = state.Patterns[patNum]
			if row, ok := pattern.GetRow(state.CurrentRow).(*Row); ok {
				return row
			}
			return nil
		}
	}
}

// GetRows returns all the rows in the pattern
func (state *PatternState) GetRows() []*Row {
	var patNum = state.GetPatNum()
	switch patNum {
	case InvalidPattern:
		return nil
	case NextPattern:
		{
			state.NextRow()
			return state.GetRows()
		}
	default:
		{
			var pattern = state.Patterns[patNum]
			pr := pattern.GetRows()

			rows := make([]*Row, len(pr))
			for i, prr := range pr {
				if r, ok := prr.(*Row); ok {
					rows[i] = r
				}
			}
			return rows
		}
	}
}
