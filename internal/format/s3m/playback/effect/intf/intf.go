package intf

// S3M is an interface to S3M effect operations
type S3M interface {
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
