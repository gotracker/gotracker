package pattern

import (
	"errors"

	"gotracker/internal/format/xm/layout"
	"gotracker/internal/player/intf"
)

// Row is a specification of the current row data
type Row struct {
	intf.Row
	Channels [32]intf.ChannelData
}

type patternLoop struct {
	Enabled bool
	Start   intf.RowIdx
	End     intf.RowIdx
	Total   uint8

	Count uint8
}

func (pl *patternLoop) ContinueLoop(currentRow intf.RowIdx) (intf.RowIdx, bool) {
	if pl.Enabled {
		if currentRow == pl.End {
			if pl.Count >= pl.Total {
				pl.Enabled = false
			} else {
				pl.Count++
				return pl.Start, true
			}
		}
	}
	return 0, false
}

func (pl *patternLoop) CommitTransaction(txn *RowUpdateTransaction) {
	if !pl.Enabled {
		if txn.patternLoopCountSet {
			pl.Enabled = true
			pl.Total = uint8(txn.patternLoopCount)
			pl.Count = 0
		}

		if txn.patternLoopStartRowIdxSet {
			pl.Start = txn.patternLoopStartRowIdx
		}

		if txn.patternLoopEndRowIdxSet {
			pl.End = txn.patternLoopEndRowIdx
		}
	}
}

// State is the current pattern state
type State struct {
	currentOrder       intf.OrderIdx
	currentRow         intf.RowIdx
	ticks              int
	tempo              int
	rowHasPatternDelay bool
	patternDelay       int
	finePatternDelay   int

	patternLoop patternLoop

	OrderLoopEnabled bool
	playedOrders     []intf.OrderIdx // when OrderLoopEnabled is false, this is used to detect loops

	Patterns []layout.Pattern
	Orders   []intf.PatternIdx
}

// GetTempo returns the tempo of the current state
func (state *State) GetTempo() int {
	return state.tempo
}

// GetSpeed returns the row speed of the current state
func (state *State) GetSpeed() int {
	return state.ticks
}

// GetTicksThisRow returns the number of ticks in the current row
func (state *State) GetTicksThisRow() int {
	rowLoops := 1
	if state.rowHasPatternDelay {
		rowLoops = state.patternDelay
	}
	extraTicks := state.finePatternDelay

	ticksThisRow := state.ticks*rowLoops + extraTicks
	return ticksThisRow
}

// GetPatNum returns the current pattern number
func (state *State) GetPatNum() intf.PatternIdx {
	if int(state.currentOrder) >= len(state.Orders) {
		return intf.InvalidPattern
	}
	return state.Orders[state.currentOrder]
}

// GetNumRows returns the number of rows in the current pattern
func (state *State) GetNumRows() int {
	if rows := state.GetRows(); rows != nil {
		return rows.NumRows()
	}
	return 0
}

// WantsStop returns true when the current pattern wants to end the song
func (state *State) WantsStop() bool {
	if state.GetPatNum() == intf.InvalidPattern {
		return true
	}
	return false
}

// setCurrentOrder sets the current order index
func (state *State) setCurrentOrder(order intf.OrderIdx) {
	prevOrder := state.currentOrder
	state.currentOrder = order
	if !state.OrderLoopEnabled && prevOrder != state.currentOrder {
		state.playedOrders = append(state.playedOrders, prevOrder)
	}
}

// GetCurrentOrder returns the current order
func (state *State) GetCurrentOrder() intf.OrderIdx {
	return state.currentOrder
}

// GetNumOrders returns the number of orders in the song
func (state *State) GetNumOrders() int {
	return len(state.Orders)
}

func (state *State) getCurrentPattern() (*layout.Pattern, error) {
	patIdx, err := state.GetCurrentPatternIdx()
	if err != nil {
		return nil, err
	}

	if int(patIdx) >= len(state.Patterns) {
		return nil, errors.New("invalid pattern index")
	}
	return &state.Patterns[patIdx], nil
}

// GetCurrentPatternIdx returns the current pattern index, derived from the order list
func (state *State) GetCurrentPatternIdx() (intf.PatternIdx, error) {
	ordLen := len(state.Orders)

	if ordLen == 0 {
		// nothing to play, don't even try
		return 0, intf.ErrStopSong
	}

	for loopCount := 0; loopCount < ordLen; loopCount++ {
		ordIdx := int(state.GetCurrentOrder())
		if ordIdx >= ordLen {
			if !state.OrderLoopEnabled {
				return 0, intf.ErrStopSong
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

		if !state.OrderLoopEnabled {
			for _, o := range state.playedOrders {
				if o == intf.OrderIdx(ordIdx) {
					return 0, intf.ErrStopSong
				}
			}
		}

		return patIdx, nil
	}
	return 0, errors.New("infinite loop detected in order list")
}

// GetCurrentRow returns the current row
func (state *State) GetCurrentRow() intf.RowIdx {
	return state.currentRow
}

// setCurrentRow sets the current row
func (state *State) setCurrentRow(row intf.RowIdx) {
	state.currentRow = row
	numRows := state.GetNumRows()
	if int(state.GetCurrentRow()) >= numRows {
		state.nextOrder(true)
	}
}

// nextOrder travels to the next pattern in the order list
func (state *State) nextOrder(resetRow ...bool) {
	state.setCurrentOrder(state.currentOrder + 1)
	state.patternLoop.Enabled = false
	state.rowHasPatternDelay = false
	state.patternDelay = 0
	state.finePatternDelay = 0
	state.GetCurrentPatternIdx() // called only to clean up order position effectparameter
	if len(resetRow) > 0 && resetRow[0] {
		state.currentRow = 0
	}
}

// Reset resets a pattern state back to zeroes
func (state *State) Reset() {
	*state = State{
		OrderLoopEnabled: true,
		playedOrders:     make([]intf.OrderIdx, 0),
	}
}

// nextRow travels to the next row in the pattern
// or the next order in the order list if the last row has been exhausted
func (state *State) nextRow() {
	if row, ok := state.patternLoop.ContinueLoop(state.GetCurrentRow()); ok {
		state.setCurrentRow(row)
	}

	state.rowHasPatternDelay = false
	state.patternDelay = 0
	state.finePatternDelay = 0

	var patNum = state.GetPatNum()
	if patNum == intf.InvalidPattern {
		return
	}

	if patNum == intf.NextPattern {
		state.nextOrder(true)
		return
	}

	nextRow := int(state.currentRow) + 1
	if nextRow >= state.GetNumRows() {
		state.nextOrder(true)
	} else {
		state.currentRow++
	}
}

// GetRow returns the current row
func (state *State) GetRow() *Row {
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
func (state *State) GetRows() intf.Rows {
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
		if int(patNum) >= len(state.Patterns) {
			return nil
		}
		pattern := state.Patterns[patNum]
		return pattern.GetRows()
	}
}

// CommitTransaction will update the order and row indexes at once, idempotently, from a row update transaction.
func (state *State) CommitTransaction(txn *RowUpdateTransaction) {
	if txn.committed {
		return
	}
	txn.committed = true

	if txn.tempoSet || txn.tempoDeltaSet {
		newTempo := state.tempo
		if txn.tempoSet {
			newTempo = txn.tempo
		}
		if txn.tempoDeltaSet {
			newTempo += txn.tempoDelta
		}
		state.tempo = newTempo
	}

	if txn.ticksSet {
		state.ticks = txn.ticks
	}

	if txn.finePatternDelaySet {
		state.finePatternDelay = txn.finePatternDelay
	}

	if !state.rowHasPatternDelay && txn.patternDelaySet {
		state.patternDelay = txn.patternDelay
		state.rowHasPatternDelay = true
	}

	state.patternLoop.CommitTransaction(txn)

	if txn.orderIdxSet || txn.rowIdxSet {
		nextRow := intf.RowIdx(0)
		if txn.rowIdxSet {
			nextRow = txn.rowIdx
		}
		if txn.orderIdxSet {
			state.setCurrentOrder(txn.orderIdx)
		} else {
			state.nextOrder(true)
		}
		state.setCurrentRow(nextRow)
	} else if txn.breakOrder {
		state.nextOrder(true)
	} else if txn.advanceRow {
		state.nextRow()
	}
}

// StartTransaction starts a row update transaction
func (state *State) StartTransaction() *RowUpdateTransaction {
	txn := RowUpdateTransaction{
		state: state,
	}

	return &txn
}
