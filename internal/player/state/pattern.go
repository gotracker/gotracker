package state

import (
	"gotracker/internal/player/intf"
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
	PatternLoopEnabled bool
	currentOrder       intf.OrderIdx
	CurrentRow         intf.RowIdx
	PlayedOrders       []intf.OrderIdx // when PatternLoopEnabled is false, this is used to detect loops

	Row RowSettings

	RowHasPatternDelay bool
	PatternDelay       int
	FinePatternDelay   int

	Patterns intf.Patterns
	Orders   []intf.PatternIdx

	LoopStart   intf.RowIdx
	LoopEnd     intf.RowIdx
	LoopTotal   uint8
	LoopEnabled bool
	LoopCount   uint8
}

// GetPatNum returns the current pattern number
func (state *PatternState) GetPatNum() intf.PatternIdx {
	if int(state.currentOrder) > len(state.Orders) {
		return intf.InvalidPattern
	}
	return state.Orders[state.currentOrder]
}

// GetNumRows returns the number of rows in the current pattern
func (state *PatternState) GetNumRows() uint8 {
	rows := state.GetRows()
	return uint8(len(rows))
}

// WantsStop returns true when the current pattern wants to end the song
func (state *PatternState) WantsStop() bool {
	if state.GetPatNum() == intf.InvalidPattern {
		return true
	}
	return false
}

// SetCurrentOrder sets the current order index
func (state *PatternState) SetCurrentOrder(order intf.OrderIdx) {
	prevOrder := state.currentOrder
	state.currentOrder = order
	if !state.PatternLoopEnabled && prevOrder != state.currentOrder {
		state.PlayedOrders = append(state.PlayedOrders, prevOrder)
	}
}

// GetCurrentOrder returns the current order
func (state *PatternState) GetCurrentOrder() intf.OrderIdx {
	return state.currentOrder
}

// NextOrder travels to the next pattern in the order list
func (state *PatternState) NextOrder(resetRow ...bool) {
	state.SetCurrentOrder(state.currentOrder + 1)
	if len(resetRow) > 0 && resetRow[0] {
		state.CurrentRow = 0
	}
}

// NextRow travels to the next row in the pattern
// or the next order in the order list if the last row has been exhausted
func (state *PatternState) NextRow() {
	var patNum = state.GetPatNum()
	if patNum == intf.InvalidPattern {
		return
	}

	if patNum == intf.NextPattern {
		state.NextOrder(true)
		return
	}

	state.CurrentRow++
	if state.CurrentRow >= intf.RowIdx(state.GetNumRows()) {
		state.NextOrder(true)
		return
	}
}

// GetRow returns the current row
func (state *PatternState) GetRow() *Row {
	var patNum = state.GetPatNum()
	switch patNum {
	case intf.InvalidPattern:
		return nil
	case intf.NextPattern:
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
	case intf.InvalidPattern:
		return nil
	case intf.NextPattern:
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
