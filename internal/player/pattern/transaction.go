package pattern

import (
	"gotracker/internal/optional"
	"gotracker/internal/player/intf"
)

type WhoJumpedFirst uint8

const (
	WhoJumpedFirstNone = WhoJumpedFirst(iota)
	WhoJumpedFirstOrder
	WhoJumpedFirstRow
)

// RowUpdateTransaction is a transactional operation for row/order updates
type RowUpdateTransaction struct {
	committed         bool
	CommitTransaction func(*RowUpdateTransaction)

	orderIdx         optional.Value //intf.OrderIdx
	rowIdx           optional.Value //intf.RowIdx
	patternDelay     optional.Value //int
	FinePatternDelay optional.Value //int
	Tempo            optional.Value //int
	Ticks            optional.Value //int
	TempoDelta       optional.Value //int

	RowIdxAllowBacktrack bool
	WhoJumpedFirst       WhoJumpedFirst
	AdvanceRow           bool
	BreakOrder           bool
}

// Cancel will mark a transaction as void/spent, i.e.: cancelled
func (txn *RowUpdateTransaction) Cancel() {
	txn.committed = true
}

// Commit will update the order and row indexes at once, idempotently.
func (txn *RowUpdateTransaction) Commit() {
	if txn.committed {
		return
	}
	if txn.CommitTransaction == nil {
		panic("cannot commit transaction using unset commit function")
	}
	txn.committed = true
	txn.CommitTransaction(txn)
}

// SetNextOrder will set the next order index
func (txn *RowUpdateTransaction) SetNextOrder(ordIdx intf.OrderIdx) {
	if !txn.orderIdx.IsSet() {
		txn.orderIdx.Set(ordIdx)
		if txn.WhoJumpedFirst == WhoJumpedFirstNone {
			txn.WhoJumpedFirst = WhoJumpedFirstOrder
		}
	}
}

// GetOrderIdx gets the order index and a flag for if it is valid/set
func (txn *RowUpdateTransaction) GetOrderIdx() (intf.OrderIdx, bool) {
	return txn.orderIdx.GetOrderIdx()
}

// SetNextRow will set the next row index
func (txn *RowUpdateTransaction) SetNextRow(rowIdx intf.RowIdx, opts ...bool) {
	if !txn.rowIdx.IsSet() {
		txn.rowIdx.Set(rowIdx)
		if txn.WhoJumpedFirst == WhoJumpedFirstNone {
			txn.WhoJumpedFirst = WhoJumpedFirstRow
		}
		if len(opts) > 0 {
			txn.RowIdxAllowBacktrack = opts[0]
		}
	}
}

// GetOrderIdx gets the row index and a flag for if it is valid/set
func (txn *RowUpdateTransaction) GetRowIdx() (intf.RowIdx, bool) {
	return txn.rowIdx.GetRowIdx()
}

// SetPatternDelay sets the row pattern delay
func (txn *RowUpdateTransaction) SetPatternDelay(patternDelay int) {
	if !txn.patternDelay.IsSet() {
		txn.patternDelay.Set(patternDelay)
	}
}

// GetPatternDelay gets the row pattern delay and a flag for if it is valid/set
func (txn *RowUpdateTransaction) GetPatternDelay() (int, bool) {
	return txn.patternDelay.GetInt()
}

// AccTempoDelta accumulates the amount of tempo delta
func (txn *RowUpdateTransaction) AccTempoDelta(delta int) {
	tempoDelta := delta
	if d, ok := txn.TempoDelta.GetInt(); ok {
		tempoDelta += d
	}
	txn.TempoDelta.Set(tempoDelta)
}
