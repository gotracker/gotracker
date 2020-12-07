package intf

type SharedMemory interface {
	SetEffectSharedMemoryIfNonZero(uint8)
	GetEffectSharedMemory(uint8) uint8
}

type Memory interface {
	PortaToNote(uint8) uint8
	Vibrato(uint8) uint8
	SampleOffset(uint8) uint8
	TempoDecrease(uint8) uint8
	TempoIncrease(uint8) uint8
}