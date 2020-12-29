package intf

// XM is an interface to XM effect operations
type XM interface {
	SetFilterEnable(bool)
	SetTicks(int)
	AddRowTicks(int)
	SetPatternDelay(int)
	SetPatternLoopStart()
	SetPatternLoopEnd()
	SetPatternLoopCount(int)
	SetTempo(int)
	DecreaseTempo(int)
	IncreaseTempo(int)
}
