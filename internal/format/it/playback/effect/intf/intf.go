package intf

// IT is an interface to IT effect operations
type IT interface {
	SetFilterEnable(bool)
	SetTicks(int)
	AddRowTicks(int)
	SetPatternDelay(int)
	SetTempo(int)
	DecreaseTempo(int)
	IncreaseTempo(int)
	SetEnvelopePosition(uint8)
}
