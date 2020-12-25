package pattern

import (
	"errors"

	"gotracker/internal/player/intf"
)

var (
	// ErrStopSong is a magic error asking to stop the current song
	ErrStopSong = errors.New("stop song")
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
	playedOrders       []intf.OrderIdx // when PatternLoopEnabled is false, this is used to detect loops

	row RowSettings

	rowHasPatternDelay bool
	patternDelay       int
	finePatternDelay   int

	Patterns intf.Patterns
	Orders   []intf.PatternIdx

	loopStart   intf.RowIdx
	loopEnd     intf.RowIdx
	loopTotal   uint8
	loopEnabled bool
	loopCount   uint8
}

// GetTempo returns the tempo of the current state
func (state *PatternState) GetTempo() int {
	return state.row.Tempo
}

// GetSpeed returns the row speed of the current state
func (state *PatternState) GetSpeed() int {
	return state.row.Ticks
}

// GetTicksThisRow returns the number of ticks in the current row
func (state *PatternState) GetTicksThisRow() int {
	rowLoops := 1
	if state.rowHasPatternDelay {
		rowLoops = state.patternDelay
	}
	extraTicks := state.finePatternDelay

	ticksThisRow := int(state.row.Ticks)*rowLoops + extraTicks
	return ticksThisRow
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
		state.playedOrders = append(state.playedOrders, prevOrder)
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
			for _, o := range state.playedOrders {
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
	state.rowHasPatternDelay = false
	state.patternDelay = 0
	state.finePatternDelay = 0
	state.GetCurrentPatternIdx() // called only to clean up order position info
	if len(resetRow) > 0 && resetRow[0] {
		state.currentRow = 0
	}
}

// Reset resets a pattern state back to zeroes
func (state *PatternState) Reset() {
	*state = PatternState{
		PatternLoopEnabled: true,
		playedOrders:       make([]intf.OrderIdx, 0),
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

// CommitTransaction will update the order and row indexes at once, idempotently, from a row update transaction.
func (state *PatternState) CommitTransaction(txn *RowUpdateTransaction) {
	if txn.committed {
		return
	}
	txn.committed = true

	if txn.tempoSet || txn.tempoDeltaSet {
		newTempo := state.row.Tempo
		if txn.tempoSet {
			newTempo = txn.tempo
		}
		if txn.tempoDeltaSet {
			newTempo += txn.tempoDelta
		}
		state.row.Tempo = newTempo
	}

	if txn.ticksSet {
		state.row.Ticks = txn.ticks
	}

	if txn.finePatternDelaySet {
		state.finePatternDelay = txn.finePatternDelay
	}

	if !state.rowHasPatternDelay && txn.patternDelaySet {
		state.patternDelay = txn.patternDelay
		state.rowHasPatternDelay = true
	}

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

// StartTransaction starts a row update transaction
func (state *PatternState) StartTransaction() *RowUpdateTransaction {
	txn := RowUpdateTransaction{
		state: state,
	}

	return &txn
}
