package intf

import (
	"time"

	"github.com/gotracker/gomixing/volume"
	device "github.com/gotracker/gosound"
	"github.com/gotracker/voice/render"

	"gotracker/internal/player/feature"
	"gotracker/internal/song"
	"gotracker/internal/song/index"
	"gotracker/internal/song/pattern"
)

// Playback is an interface for rendering a song to output data
type Playback interface {
	Update(time.Duration, chan<- *device.PremixData) error
	Generate(time.Duration) (*device.PremixData, error)

	GetSongData() song.Data

	GetNumChannels() int
	GetNumOrders() int
	SetNextOrder(index.Order) error
	SetNextRow(index.Row, ...bool) error
	GetCurrentRow() index.Row
	GetGlobalVolume() volume.Volume
	SetGlobalVolume(volume.Volume)
	Configure([]feature.Feature)
	GetName() string
	CanOrderLoop() bool
	BreakOrder() error
	SetOnEffect(func(Effect))
	IgnoreUnknownEffect() bool

	StartPatternTransaction() *pattern.RowUpdateTransaction

	SetupSampler(int, int, int) error
	GetSampleRate() float32
	GetOPL2Chip() render.OPL2Chip
}
