package intf

// XM is an interface to XM effect operations
type XM interface {
	SetFilterEnable(bool)
	SetTicks(int)
	AddRowTicks(int)
	SetPatternDelay(int)
	SetTempo(int)
	DecreaseTempo(int)
	IncreaseTempo(int)
	SetEnvelopePosition(uint8)
}
