package intf

import "gotracker/internal/player/intf"

// S3M is an interface to S3M effect operations
type S3M interface {
	SetFilterEnable(bool)
	SetTicks(int)
	AddRowTicks(int)
	SetPatternDelay(int)
	SetPatternLoopStart(intf.RowIdx)
	SetPatternLoopEnd()
	SetPatternLoopCount(int)
	SetTempo(int)
	DecreaseTempo(int)
	IncreaseTempo(int)
}
