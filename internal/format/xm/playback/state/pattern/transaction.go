package pattern

import "gotracker/internal/player/intf"

type whoJumpedFirst uint8

const (
	wjfNone = whoJumpedFirst(iota)
	wjfOrder
	wjfRow
)

// RowUpdateTransaction is a transactional operation for row/order updates
type RowUpdateTransaction struct {
	intf.SongPositionState
	orderIdx                  intf.OrderIdx
	orderIdxSet               bool
	rowIdx                    intf.RowIdx
	rowIdxSet                 bool
	whoJumpedFirst            whoJumpedFirst
	advanceRow                bool
	breakOrder                bool
	committed                 bool
	patternLoopStartRowIdx    intf.RowIdx
	patternLoopStartRowIdxSet bool
	patternLoopEndRowIdx      intf.RowIdx
	patternLoopEndRowIdxSet   bool
	patternLoopCount          int
	patternLoopCountSet       bool
	rowHasPatternDelay        bool
	patternDelay              int
	patternDelaySet           bool
	finePatternDelay          int
	finePatternDelaySet       bool
	tempo                     int
	tempoSet                  bool
	ticks                     int
	ticksSet                  bool
	tempoDelta                int
	tempoDeltaSet             bool
	state                     *State
}

// Cancel will mark a transaction as void/spent, i.e.: cancelled
func (txn *RowUpdateTransaction) Cancel() {
	txn.committed = true
}

// Commit will update the order and row indexes at once, idempotently.
func (txn *RowUpdateTransaction) Commit() {
	txn.state.CommitTransaction(txn)
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
	if !txn.orderIdxSet {
		txn.orderIdx = ordIdx
		txn.orderIdxSet = true
		if txn.whoJumpedFirst == wjfNone {
			txn.whoJumpedFirst = wjfOrder
		}
	}
}

// SetNextRow will set the next row index
func (txn *RowUpdateTransaction) SetNextRow(rowIdx intf.RowIdx) {
	if !txn.rowIdxSet {
		txn.rowIdx = rowIdx
		txn.rowIdxSet = true
		if txn.whoJumpedFirst == wjfNone {
			txn.whoJumpedFirst = wjfRow
		}
	}
}

// SetPatternLoopStart will set the pattern loop starting row index
func (txn *RowUpdateTransaction) SetPatternLoopStart(row intf.RowIdx) {
	txn.patternLoopStartRowIdx = row
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

// SetPatternDelay sets the row pattern delay
func (txn *RowUpdateTransaction) SetPatternDelay(patternDelay int) {
	if !txn.rowHasPatternDelay {
		txn.rowHasPatternDelay = true

		txn.patternDelay = patternDelay
		txn.patternDelaySet = true
	}
}

// SetTempo will set the row tempo
func (txn *RowUpdateTransaction) SetTempo(tempo int) {
	txn.tempo = tempo
	txn.tempoSet = true
}

// SetTicks will set the row ticks
func (txn *RowUpdateTransaction) SetTicks(ticks int) {
	txn.ticks = ticks
	txn.ticksSet = true
}

// AccTempoDelta accumulates the amount of tempo delta
func (txn *RowUpdateTransaction) AccTempoDelta(delta int) {
	txn.tempoDelta += delta
	txn.tempoDeltaSet = true
}

// SetFinePatternDelay will set the fine pattern delay row ticks
func (txn *RowUpdateTransaction) SetFinePatternDelay(ticks int) {
	txn.finePatternDelay = ticks
	txn.finePatternDelaySet = true
}
