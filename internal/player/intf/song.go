package intf

import "gotracker/internal/s3m/volume"

type Song interface {
	SetCurrentOrder(uint8)
	SetCurrentRow(uint8)
	SetTempo(int)
	DecreaseTempo(int)
	IncreaseTempo(int)
	SetGlobalVolume(volume.Volume)
	SetTicks(int)
	AddRowTicks(int)
	SetPatternDelay(int)
	SetPatternLoopStart()
	SetPatternLoopEnd(uint8)
}
