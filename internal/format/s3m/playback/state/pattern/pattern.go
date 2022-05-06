package pattern

import (
	"errors"

	formatutil "github.com/gotracker/gotracker/internal/format/internal/util"
	"github.com/gotracker/gotracker/internal/format/s3m/layout/channel"
	"github.com/gotracker/gotracker/internal/optional"
	"github.com/gotracker/gotracker/internal/player/feature"
	"github.com/gotracker/gotracker/internal/song"
	"github.com/gotracker/gotracker/internal/song/index"
	"github.com/gotracker/gotracker/internal/song/pattern"
)

// State is the current pattern state
type State struct {
	currentOrder      index.Order
	currentRow        index.Row
	ticks             int
	tempo             int
	patternDelay      optional.Value[int]
	finePatternDelay  int
	resetPatternLoops bool

	SongLoop             feature.SongLoop
	PlayUntilOrderAndRow feature.PlayUntilOrderAndRow
	loopDetect           formatutil.LoopDetect // when SongLoopEnabled is false, this is used to detect song loops
	loopCount            int

	Patterns []pattern.Pattern[channel.Data]
	Orders   []index.Pattern
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
	if patternDelay, ok := state.patternDelay.Get(); ok {
		rowLoops = patternDelay
	}
	extraTicks := state.finePatternDelay

	ticksThisRow := state.ticks*rowLoops + extraTicks
	return ticksThisRow
}

// GetPatNum returns the current pattern number
func (state *State) GetPatNum() index.Pattern {
	if int(state.currentOrder) >= len(state.Orders) {
		return index.InvalidPattern
	}
	return state.Orders[state.currentOrder]
}

// GetNumRows returns the number of rows in the current pattern
func (state *State) GetNumRows() (int, error) {
	rows, err := state.GetRows()
	if err != nil {
		return 0, err
	}
	if rows != nil {
		return rows.NumRows(), nil
	}
	return 0, nil
}

// WantsStop returns true when the current pattern wants to end the song
func (state *State) WantsStop() bool {
	return state.GetPatNum() == index.InvalidPattern
}

// setCurrentOrder sets the current order index
func (state *State) setCurrentOrder(order index.Order) {
	state.currentOrder = order
	state.resetPatternLoops = true
}

func (state *State) advanceOrder() {
	state.setCurrentOrder(state.currentOrder + 1)
}

// GetCurrentOrder returns the current order
func (state *State) GetCurrentOrder() index.Order {
	return state.currentOrder
}

// GetNumOrders returns the number of orders in the song
func (state *State) GetNumOrders() int {
	return len(state.Orders)
}

// GetCurrentPatternIdx returns the current pattern index, derived from the order list
func (state *State) GetCurrentPatternIdx() (index.Pattern, error) {
	ordLen := len(state.Orders)

	if ordLen == 0 {
		// nothing to play, don't even try
		return 0, song.ErrStopSong
	}

	for loopCount := 0; loopCount < ordLen; loopCount++ {
		ordIdx := int(state.GetCurrentOrder())
		if ordIdx >= ordLen {
			if !(state.SongLoop.Count < 0 || state.loopCount < state.SongLoop.Count) {
				return 0, song.ErrStopSong
			}
			state.setCurrentOrder(0)
			continue
		}

		patIdx := state.Orders[ordIdx]
		if patIdx == index.NextPattern {
			if err := state.nextOrder(true); err != nil {
				return 0, err
			}
			continue
		}

		if patIdx == index.InvalidPattern {
			if err := state.nextOrder(true); err != nil {
				return 0, err
			}
			continue // this is supposed to be a song break
		}

		return patIdx, nil
	}
	return 0, errors.New("infinite loop detected in order list")
}

// GetCurrentRow returns the current row
func (state *State) GetCurrentRow() index.Row {
	return state.currentRow
}

// setCurrentRow sets the current row
func (state *State) setCurrentRow(row index.Row) error {
	state.currentRow = row
	rows, err := state.GetNumRows()
	if err != nil {
		return err
	}
	if int(state.GetCurrentRow()) >= rows {
		if err := state.nextOrder(true); err != nil {
			return err
		}
	}
	return nil
}

// Observe will attempt to detect a song loop
func (state *State) Observe() error {
	if state.SongLoop.Count >= 0 {
		if state.loopDetect.Observe(state.currentOrder, state.currentRow) {
			if state.SongLoop.Count > 0 && state.loopCount >= state.SongLoop.Count {
				return song.ErrStopSong
			}
			state.loopCount += 1
			state.loopDetect.Reset()
		}
	}
	if state.currentOrder == index.Order(state.PlayUntilOrderAndRow.Order) && state.currentRow == index.Row(state.PlayUntilOrderAndRow.Row) {
		if state.SongLoop.Count >= 0 && state.loopCount >= state.SongLoop.Count {
			return song.ErrStopSong
		}
	}
	return nil
}

// nextOrder travels to the next pattern in the order list
func (state *State) nextOrder(resetRow ...bool) error {
	state.advanceOrder()
	state.patternDelay.Reset()
	state.finePatternDelay = 0
	// called only to clean up order position info
	if _, err := state.GetCurrentPatternIdx(); err != nil {
		return err
	}
	if len(resetRow) > 0 && resetRow[0] {
		state.currentRow = 0
	}
	return nil
}

// Reset resets a pattern state back to zeroes
func (state *State) Reset() {
	*state = State{
		SongLoop: feature.SongLoop{
			Count: 0,
		},
		PlayUntilOrderAndRow: feature.PlayUntilOrderAndRow{
			Order: -1,
			Row:   -1,
		},
	}
}

// nextRow travels to the next row in the pattern
// or the next order in the order list if the last row has been exhausted
func (state *State) nextRow() error {
	state.patternDelay.Reset()
	state.finePatternDelay = 0

	var patNum = state.GetPatNum()
	if patNum == index.InvalidPattern {
		return nil
	}

	if patNum == index.NextPattern {
		if err := state.nextOrder(true); err != nil {
			return err
		}
		return nil
	}

	rows, err := state.GetNumRows()
	if err != nil {
		return err
	}
	if state.currentRow.Increment(rows) {
		if err := state.nextOrder(true); err != nil {
			return err
		}
	}
	return nil
}

// GetRows returns all the rows in the pattern
func (state *State) GetRows() (song.Rows[channel.Data], error) {
nextRow:
	for loops := 0; loops < len(state.Patterns); loops++ {
		var patNum = state.GetPatNum()
		switch patNum {
		case index.InvalidPattern:
			return nil, nil
		case index.NextPattern:
			if err := state.nextRow(); err != nil {
				return nil, err
			}
			continue nextRow
		default:
			if int(patNum) >= len(state.Patterns) {
				return nil, nil
			}
			pattern := state.Patterns[patNum]
			return pattern.GetRows(), nil
		}
	}
	return nil, nil
}

// NeedResetPatternLoops returns the state of the resetPatternLoops variable (and resets it)
func (state *State) NeedResetPatternLoops() bool {
	rpl := state.resetPatternLoops
	state.resetPatternLoops = false
	return rpl
}

// commitTransaction will update the order and row indexes at once, idempotently, from a row update transaction.
func (state *State) commitTransaction(txn *pattern.RowUpdateTransaction) error {
	tempo, tempoSet := txn.Tempo.Get()
	tempoDelta, tempoDeltaSet := txn.TempoDelta.Get()
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

	if ticks, ok := txn.Ticks.Get(); ok {
		state.ticks = ticks
	}

	if finePatternDelay, ok := txn.FinePatternDelay.Get(); ok {
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
				if err := state.setCurrentRow(0); err != nil {
					return err
				}
			}
		}
		if rowIdxSet {
			if !orderIdxSet && !txn.RowIdxAllowBacktrack { //  && state.currentRow > rowIdx   // QUIRK[S3M/MOD]
				if err := state.nextOrder(); err != nil {
					return err
				}
			}
			if err := state.setCurrentRow(rowIdx); err != nil {
				return err
			}
		}
	} else if txn.BreakOrder { // QUIRK[S3M/MOD]
		if err := state.nextOrder(true); err != nil {
			return err
		}
	} else if txn.AdvanceRow {
		if err := state.nextRow(); err != nil {
			return err
		}
	}
	return nil
}

// StartTransaction starts a row update transaction
func (state *State) StartTransaction() *pattern.RowUpdateTransaction {
	txn := pattern.RowUpdateTransaction{
		CommitTransaction: state.commitTransaction,
	}

	return &txn
}
