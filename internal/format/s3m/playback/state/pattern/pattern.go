package pattern

import (
	"errors"

	formatutil "gotracker/internal/format/internal/util"
	"gotracker/internal/optional"
	"gotracker/internal/player/intf"
	"gotracker/internal/player/pattern"
)

// State is the current pattern state
type State struct {
	currentOrder      intf.OrderIdx
	currentRow        intf.RowIdx
	ticks             int
	tempo             int
	patternDelay      optional.Value //int
	finePatternDelay  int
	resetPatternLoops bool

	SongLoopEnabled bool
	loopDetect      formatutil.LoopDetect // when SongLoopEnabled is false, this is used to detect song loops

	Patterns []pattern.Pattern
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
	if patternDelay, ok := state.patternDelay.GetInt(); ok {
		rowLoops = patternDelay
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
	return state.GetPatNum() == intf.InvalidPattern
}

// setCurrentOrder sets the current order index
func (state *State) setCurrentOrder(order intf.OrderIdx) {
	state.currentOrder = order
	state.resetPatternLoops = true
}

func (state *State) advanceOrder() {
	state.setCurrentOrder(state.currentOrder + 1)
}

// GetCurrentOrder returns the current order
func (state *State) GetCurrentOrder() intf.OrderIdx {
	return state.currentOrder
}

// GetNumOrders returns the number of orders in the song
func (state *State) GetNumOrders() int {
	return len(state.Orders)
}

// NeedResetPatternLoops returns the state of the resetPatternLoops variable (and resets it)
func (state *State) NeedResetPatternLoops() bool {
	rpl := state.resetPatternLoops
	state.resetPatternLoops = false
	return rpl
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
			if !state.SongLoopEnabled {
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
	if int(state.GetCurrentRow()) >= state.GetNumRows() {
		state.nextOrder(true)
	}
}

// Observe will attempt to detect a song loop
func (state *State) Observe() error {
	if !state.SongLoopEnabled && state.loopDetect.Observe(state.currentOrder, state.currentRow) {
		return intf.ErrStopSong
	}
	return nil
}

// nextOrder travels to the next pattern in the order list
func (state *State) nextOrder(resetRow ...bool) {
	state.advanceOrder()
	state.patternDelay.Reset()
	state.finePatternDelay = 0
	_, _ = state.GetCurrentPatternIdx() // called only to clean up order position info
	if len(resetRow) > 0 && resetRow[0] {
		state.currentRow = 0
	}
}

// Reset resets a pattern state back to zeroes
func (state *State) Reset() {
	*state = State{
		SongLoopEnabled: true,
	}
}

// nextRow travels to the next row in the pattern
// or the next order in the order list if the last row has been exhausted
func (state *State) nextRow() {
	state.patternDelay.Reset()
	state.finePatternDelay = 0

	var patNum = state.GetPatNum()
	if patNum == intf.InvalidPattern {
		return
	}

	if patNum == intf.NextPattern {
		state.nextOrder(true)
		return
	}

	if state.currentRow.Increment(state.GetNumRows()) {
		state.nextOrder(true)
	}
}

// GetRows returns all the rows in the pattern
func (state *State) GetRows() intf.Rows {
nextRow:
	for loops := 0; loops < len(state.Patterns); loops++ {
		var patNum = state.GetPatNum()
		switch patNum {
		case intf.InvalidPattern:
			return nil
		case intf.NextPattern:
			state.nextRow()
			continue nextRow
		default:
			if int(patNum) >= len(state.Patterns) {
				return nil
			}
			pattern := state.Patterns[patNum]
			return pattern.GetRows()
		}
	}
	return nil
}

// commitTransaction will update the order and row indexes at once, idempotently, from a row update transaction.
func (state *State) commitTransaction(txn *pattern.RowUpdateTransaction) {
	tempo, tempoSet := txn.Tempo.GetInt()
	tempoDelta, tempoDeltaSet := txn.TempoDelta.GetInt()
	if tempoSet || tempoDeltaSet {
		newTempo := state.tempo
		if tempoSet {
			newTempo = tempo
		}
		if tempoDeltaSet {
			newTempo += tempoDelta
		}
		state.tempo = newTempo
	}

	if ticks, ok := txn.Ticks.GetInt(); ok {
		state.ticks = ticks
	}

	if finePatternDelay, ok := txn.FinePatternDelay.GetInt(); ok {
		state.finePatternDelay = finePatternDelay
	}

	if !state.patternDelay.IsSet() {
		if patternDelay, ok := txn.GetPatternDelay(); ok {
			state.patternDelay.Set(patternDelay)
		}
	}

	orderIdx, orderIdxSet := txn.GetOrderIdx()
	rowIdx, rowIdxSet := txn.GetRowIdx()

	if orderIdxSet || rowIdxSet {
		if orderIdxSet {
			state.setCurrentOrder(orderIdx)
			if !rowIdxSet {
				state.setCurrentRow(0)
			}
		}
		if rowIdxSet {
			if !orderIdxSet && !txn.RowIdxAllowBacktrack {
				state.nextOrder()
			}
			state.setCurrentRow(rowIdx)
		}
	} else if txn.BreakOrder {
		state.nextOrder(true)
	} else if txn.AdvanceRow {
		state.nextRow()
	}
}

// StartTransaction starts a row update transaction
func (state *State) StartTransaction() *pattern.RowUpdateTransaction {
	txn := pattern.RowUpdateTransaction{
		CommitTransaction: state.commitTransaction,
	}

	return &txn
}
