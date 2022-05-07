package intf

import (
	"github.com/gotracker/gomixing/volume"
	"github.com/gotracker/gotracker/internal/song/index"
)

// S3M is an interface to S3M effect operations
type S3M interface {
	SetTicks(int) error                            // Axx
	SetNextOrder(index.Order) error                // Bxx
	SetNextRow(index.Row) error                    // Cxx
	SetFilterEnable(bool)                          // S0x
	SetNextRowWithBacktrack(index.Row, bool) error // SBx
	GetCurrentRow() index.Row                      // SBx
	SetPatternDelay(int) error                     // SEx
	AddRowTicks(int) error                         // S6x
	SetTempo(int) error                            // Txx
	IncreaseTempo(int) error                       // Txx
	DecreaseTempo(int) error                       // Txx
	SetGlobalVolume(volume.Volume)                 // Vxx
	IgnoreUnknownEffect() bool                     // Unhandled
}
