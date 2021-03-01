package intf

// IT is an interface to IT effect operations
type IT interface {
	SetFilterEnable(bool)
	SetTicks(int) error
	AddRowTicks(int) error
	SetPatternDelay(int) error
	SetTempo(int) error
	DecreaseTempo(int) error
	IncreaseTempo(int) error
	SetEnvelopePosition(uint8)
}
