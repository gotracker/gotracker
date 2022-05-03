package intf

import "gotracker/internal/format/xm/layout/channel"

// XM is an interface to XM effect operations
type XM interface {
	SetFilterEnable(bool)
	SetTicks(int) error
	AddRowTicks(int) error
	SetPatternDelay(int) error
	SetTempo(int) error
	DecreaseTempo(int) error
	IncreaseTempo(int) error
	SetEnvelopePosition(channel.DataEffect)
}
