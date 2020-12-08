package intf

// SharedMemory is an interface to storing effect data on the channel state
type SharedMemory interface {
	SetEffectSharedMemoryIfNonZero(uint8)
	GetEffectSharedMemory(uint8) uint8
}

// Memory is an interface for storing effect data on the channel state for specific effects
type Memory interface {
	PortaToNote(uint8) uint8
	Vibrato(uint8) uint8
	SampleOffset(uint8) uint8
	TempoDecrease(uint8) uint8
	TempoIncrease(uint8) uint8
}
