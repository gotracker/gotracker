package state

import (
	"gotracker/internal/player/intf"
)

type PatternNum uint8

const (
	NextPattern    = PatternNum(254)
	InvalidPattern = PatternNum(255)
)

type RowSettings struct {
	Ticks int
	Tempo int
}

type Row struct {
	intf.Row
	Channels [32]intf.ChannelData
}

type PatternState struct {
	CurrentOrder uint8
	CurrentRow   uint8

	Row RowSettings

	RowHasPatternDelay bool
	PatternDelay       int
	FinePatternDelay   int

	Patterns *[]intf.Pattern
	Orders   *[]uint8

	LoopStart   uint8
	LoopEnd     uint8
	LoopTotal   uint8
	LoopEnabled bool
	LoopCount   uint8
}

func (state *PatternState) GetPatNum() PatternNum {
	if int(state.CurrentOrder) > len(*state.Orders) {
		return InvalidPattern
	}
	return PatternNum((*state.Orders)[state.CurrentOrder])
}

func (state *PatternState) WantsStop() bool {
	if state.GetPatNum() == InvalidPattern {
		return true
	}
	return false
}

func (state *PatternState) NextOrder() {
	state.CurrentOrder++
	state.CurrentRow = 0
}

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
	if state.CurrentRow >= 64 {
		state.NextOrder()
		return
	}
}

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
			var pattern = (*state.Patterns)[patNum]
			if row, ok := pattern.GetRow(state.CurrentRow).(*Row); ok {
				return row
			}
			return nil
		}
	}
}

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
			var pattern = (*state.Patterns)[patNum]
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
