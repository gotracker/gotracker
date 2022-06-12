package intf

import (
	"time"

	device "github.com/gotracker/gosound"

	"github.com/gotracker/gotracker/internal/player/feature"
	"github.com/gotracker/gotracker/internal/player/output"
	"github.com/gotracker/gotracker/internal/song"
	"github.com/gotracker/gotracker/internal/song/index"
	"github.com/gotracker/gotracker/internal/song/pattern"
)

// Playback is an interface for rendering a song to output data
type Playback interface {
	output.ConfigIntf

	Update(time.Duration, chan<- *device.PremixData) error
	Generate(time.Duration) (*device.PremixData, error)

	GetSongData() song.Data

	GetNumChannels() int
	GetNumOrders() int
	SetNextOrder(index.Order) error
	SetNextRow(index.Row) error
	SetNextRowWithBacktrack(index.Row, bool) error
	GetCurrentRow() index.Row
	Configure([]feature.Feature) error
	GetName() string
	CanOrderLoop() bool
	BreakOrder() error
	SetOnEffect(func(Effect))
	IgnoreUnknownEffect() bool

	StartPatternTransaction() *pattern.RowUpdateTransaction
}
