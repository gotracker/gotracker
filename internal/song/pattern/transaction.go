package pattern

import (
	"gotracker/internal/optional"
	"gotracker/internal/song/index"
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
	CommitTransaction func(*RowUpdateTransaction) error

	orderIdx         optional.Value[index.Order]
	rowIdx           optional.Value[index.Row]
	patternDelay     optional.Value[int]
	FinePatternDelay optional.Value[int]
	Tempo            optional.Value[int]
	Ticks            optional.Value[int]
	TempoDelta       optional.Value[int]

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
func (txn *RowUpdateTransaction) Commit() error {
	if txn.committed {
		return nil
	}
	if txn.CommitTransaction == nil {
		panic("cannot commit transaction using unset commit function")
	}
	txn.committed = true
	return txn.CommitTransaction(txn)
}

// SetNextOrder will set the next order index
func (txn *RowUpdateTransaction) SetNextOrder(ordIdx index.Order) {
	if !txn.orderIdx.IsSet() {
		txn.orderIdx.Set(ordIdx)
		if txn.WhoJumpedFirst == WhoJumpedFirstNone {
			txn.WhoJumpedFirst = WhoJumpedFirstOrder
		}
	}
}

// GetOrderIdx gets the order index and a flag for if it is valid/set
func (txn *RowUpdateTransaction) GetOrderIdx() (index.Order, bool) {
	return txn.orderIdx.Get()
}

// SetNextRow will set the next row index
func (txn *RowUpdateTransaction) SetNextRow(rowIdx index.Row, opts ...bool) {
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
func (txn *RowUpdateTransaction) GetRowIdx() (index.Row, bool) {
	return txn.rowIdx.Get()
}

// SetPatternDelay sets the row pattern delay
func (txn *RowUpdateTransaction) SetPatternDelay(patternDelay int) {
	if !txn.patternDelay.IsSet() {
		txn.patternDelay.Set(patternDelay)
	}
}

// GetPatternDelay gets the row pattern delay and a flag for if it is valid/set
func (txn *RowUpdateTransaction) GetPatternDelay() (int, bool) {
	return txn.patternDelay.Get()
}

// AccTempoDelta accumulates the amount of tempo delta
func (txn *RowUpdateTransaction) AccTempoDelta(delta int) {
	tempoDelta := delta
	if d, ok := txn.TempoDelta.Get(); ok {
		tempoDelta += d
	}
	txn.TempoDelta.Set(tempoDelta)
}
