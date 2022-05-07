package intf

import (
	"github.com/gotracker/gomixing/volume"
	"github.com/gotracker/gotracker/internal/song/index"
)

// XM is an interface to XM effect operations
type XM interface {
	SetNextOrder(index.Order) error                // Bxx
	BreakOrder() error                             // Dxx
	SetNextRow(index.Row) error                    // Dxx
	SetNextRowWithBacktrack(index.Row, bool) error // E6x
	GetCurrentRow() index.Row                      // E6x
	SetPatternDelay(int) error                     // EEx
	SetTicks(int) error                            // Fxx
	SetTempo(int) error                            // Fxx
	SetGlobalVolume(volume.Volume)                 // Gxx
	GetGlobalVolume() volume.Volume                // Hxx
	SetEnvelopePosition(int)                       // Lxx
	IgnoreUnknownEffect() bool                     // Unhandled
}
