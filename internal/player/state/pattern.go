package state

import (
	"s3mplayer/internal/player/channel"
	"s3mplayer/internal/s3m"
)

type PatternNum uint8

const (
	NextPattern    = PatternNum(254)
	InvalidPattern = PatternNum(255)
)

type Row struct {
	Ticks int
	Tempo int
}

type PatternState struct {
	CurrentOrder uint8
	CurrentRow   uint8

	Row Row

	RowHasPatternDelay bool
	PatternDelay       int
	FinePatternDelay   int

	Patterns *[]s3m.Pattern
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

func (state *PatternState) GetRow() *[32]channel.Data {
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
			var pattern = &(*state.Patterns)[patNum]
			return &pattern.Rows[state.CurrentRow]
		}
	}
}
