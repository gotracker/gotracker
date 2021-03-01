package intf

// S3M is an interface to S3M effect operations
type S3M interface {
	SetFilterEnable(bool)
	SetTicks(int) error
	AddRowTicks(int) error
	SetPatternDelay(int) error
	SetTempo(int) error
	DecreaseTempo(int) error
	IncreaseTempo(int) error
}
