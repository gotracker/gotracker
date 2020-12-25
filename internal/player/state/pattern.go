package state

import (
	"errors"
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
	currentRow         intf.RowIdx
	PlayedOrders       []intf.OrderIdx // when PatternLoopEnabled is false, this is used to detect loops

	Row RowSettings

	RowHasPatternDelay bool
	PatternDelay       int
	FinePatternDelay   int

	Patterns intf.Patterns
	Orders   []intf.PatternIdx

	loopStart   intf.RowIdx
	loopEnd     intf.RowIdx
	loopTotal   uint8
	loopEnabled bool
	loopCount   uint8
}

// GetPatNum returns the current pattern number
func (state *PatternState) GetPatNum() intf.PatternIdx {
	if int(state.currentOrder) >= len(state.Orders) {
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

// setCurrentOrder sets the current order index
func (state *PatternState) setCurrentOrder(order intf.OrderIdx) {
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

func (state *PatternState) getCurrentPattern() (intf.Pattern, error) {
	patIdx, err := state.GetCurrentPatternIdx()
	if err != nil {
		return nil, err
	}

	if int(patIdx) >= len(state.Patterns) {
		return nil, errors.New("invalid pattern index")
	}
	return state.Patterns[patIdx], nil
}

// GetCurrentPatternIdx returns the current pattern index, derived from the order list
func (state *PatternState) GetCurrentPatternIdx() (intf.PatternIdx, error) {
	ordLen := len(state.Orders)

	if ordLen == 0 {
		// nothing to play, don't even try
		return 0, ErrStopSong
	}

	for loopCount := 0; loopCount < ordLen; loopCount++ {
		ordIdx := int(state.GetCurrentOrder())
		if ordIdx >= ordLen {
			if !state.PatternLoopEnabled {
				return 0, ErrStopSong
			}
			state.setCurrentOrder(0)
			continue
		}

		patIdx := state.Orders[ordIdx]
		if patIdx == intf.NextPattern {
			state.nextOrder(true)
			continue
		}

		if patIdx == intf.InvalidPattern {
			state.nextOrder(true)
			continue // this is supposed to be a song break
		}

		if !state.PatternLoopEnabled {
			for _, o := range state.PlayedOrders {
				if o == intf.OrderIdx(ordIdx) {
					return 0, ErrStopSong
				}
			}
		}

		return patIdx, nil
	}
	return 0, errors.New("infinite loop detected in order list")
}

// GetCurrentRow returns the current row
func (state *PatternState) GetCurrentRow() intf.RowIdx {
	return state.currentRow
}

// setCurrentRow sets the current row
func (state *PatternState) setCurrentRow(row intf.RowIdx) {
	state.currentRow = row
	if state.GetCurrentRow() >= intf.RowIdx(state.GetNumRows()) {
		state.nextOrder(true)
	}
}

// nextOrder travels to the next pattern in the order list
func (state *PatternState) nextOrder(resetRow ...bool) {
	state.setCurrentOrder(state.currentOrder + 1)
	state.loopEnabled = false
	state.GetCurrentPatternIdx() // called only to clean up order position info
	if len(resetRow) > 0 && resetRow[0] {
		state.currentRow = 0
	}
}

// Reset resets a pattern state back to zeroes
func (state *PatternState) Reset() {
	*state = PatternState{
		PatternLoopEnabled: true,
		PlayedOrders:       make([]intf.OrderIdx, 0),
	}
}

// nextRow travels to the next row in the pattern
// or the next order in the order list if the last row has been exhausted
func (state *PatternState) nextRow() {
	if state.loopEnabled {
		if state.GetCurrentRow() == state.loopEnd {
			if state.loopCount >= state.loopTotal {
				state.loopEnabled = false
			} else {
				state.loopCount++
				state.setCurrentRow(state.loopStart)
				return
			}
		}
	}

	var patNum = state.GetPatNum()
	if patNum == intf.InvalidPattern {
		return
	}

	if patNum == intf.NextPattern {
		state.nextOrder(true)
		return
	}

	state.currentRow++
	if state.currentRow >= intf.RowIdx(state.GetNumRows()) {
		state.nextOrder(true)
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
			state.nextRow()
			return state.GetRow()
		}
	default:
		{
			var pattern = state.Patterns[patNum]
			if row, ok := pattern.GetRow(state.currentRow).(*Row); ok {
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
			state.nextRow()
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

// RowUpdateTransaction is a transactional operation for row/order updates
type RowUpdateTransaction struct {
	intf.SongPositionState
	orderIdx                  intf.OrderIdx
	orderIdxSet               bool
	rowIdx                    intf.RowIdx
	rowIdxSet                 bool
	advanceRow                bool
	breakOrder                bool
	committed                 bool
	patternLoopStartRowIdx    intf.RowIdx
	patternLoopStartRowIdxSet bool
	patternLoopEndRowIdx      intf.RowIdx
	patternLoopEndRowIdxSet   bool
	patternLoopCount          int
	patternLoopCountSet       bool
	state                     *PatternState
}

// Cancel will mark a transaction as void/spent, i.e.: cancelled
func (txn *RowUpdateTransaction) Cancel() {
	txn.committed = true
}

// Commit will update the order and row indexes at once, idempotently.
func (txn *RowUpdateTransaction) Commit() {
	txn.state.CommitTransaction(txn)
}

// CommitTransaction will update the order and row indexes at once, idempotently, from a row update transaction.
func (state *PatternState) CommitTransaction(txn *RowUpdateTransaction) {
	if txn.committed {
		return
	}
	txn.committed = true

	if !state.loopEnabled {
		if txn.patternLoopCountSet {
			state.loopEnabled = true
			state.loopTotal = uint8(txn.patternLoopCount)
			state.loopCount = 0
		}

		if txn.patternLoopStartRowIdxSet {
			state.loopStart = txn.patternLoopStartRowIdx
		}

		if txn.patternLoopEndRowIdxSet {
			state.loopEnd = txn.patternLoopEndRowIdx
		}
	}

	if txn.orderIdxSet || txn.rowIdxSet {
		if txn.orderIdxSet {
			state.setCurrentOrder(txn.orderIdx)
		}
		if txn.rowIdxSet {
			if !txn.orderIdxSet && state.currentRow > txn.rowIdx {
				state.nextOrder()
			}
			state.setCurrentRow(txn.rowIdx)
		}
	} else if txn.breakOrder {
		state.nextOrder(true)
	} else if txn.advanceRow {
		state.nextRow()
	}
}

// AdvanceRow will advance the row index, which might also advance the order index
func (txn *RowUpdateTransaction) AdvanceRow() {
	txn.advanceRow = true
}

// BreakOrder will advance to the next order index and reset the row index to 0
func (txn *RowUpdateTransaction) BreakOrder() {
	txn.breakOrder = true
}

// SetNextOrder will set the next order index
func (txn *RowUpdateTransaction) SetNextOrder(ordIdx intf.OrderIdx) {
	txn.orderIdx = ordIdx
	txn.orderIdxSet = true
}

// SetNextRow will set the next row index
func (txn *RowUpdateTransaction) SetNextRow(rowIdx intf.RowIdx) {
	txn.rowIdx = rowIdx
	txn.rowIdxSet = true
}

// SetPatternLoopStart will set the pattern loop starting row index
func (txn *RowUpdateTransaction) SetPatternLoopStart() {
	txn.patternLoopStartRowIdx = txn.state.currentRow
	txn.patternLoopStartRowIdxSet = true
}

// SetPatternLoopEnd will set the pattern loop ending row index
func (txn *RowUpdateTransaction) SetPatternLoopEnd() {
	txn.patternLoopEndRowIdx = txn.state.currentRow
	txn.patternLoopEndRowIdxSet = true
}

// SetPatternLoopCount will set the pattern loop ending row index
func (txn *RowUpdateTransaction) SetPatternLoopCount(count int) {
	txn.patternLoopCount = count
	txn.patternLoopCountSet = true
}

// StartTransaction starts a row update transaction
func (state *PatternState) StartTransaction() *RowUpdateTransaction {
	txn := RowUpdateTransaction{
		state: state,
	}

	return &txn
}
