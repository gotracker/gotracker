package intf

import (
	"github.com/gotracker/gomixing/volume"
	"github.com/gotracker/gotracker/internal/song/index"
)

// IT is an interface to IT effect operations
type IT interface {
	SetTicks(int) error                            // Axx
	SetNextOrder(index.Order) error                // Bxx
	SetNextRow(index.Row) error                    // Cxx
	AddRowTicks(int) error                         // S6x
	SetNextRowWithBacktrack(index.Row, bool) error // SBx
	GetCurrentRow() index.Row                      // SBx
	SetPatternDelay(int) error                     // SEx
	SetTempo(int) error                            // Txx
	IncreaseTempo(int) error                       // Txx
	DecreaseTempo(int) error                       // Txx
	SetGlobalVolume(volume.Volume)                 // Vxx, Wxx
	GetGlobalVolume() volume.Volume                // Vxx, Wxx
	IgnoreUnknownEffect() bool                     // Unhandled
}
