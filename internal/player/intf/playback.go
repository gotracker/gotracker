package intf

import (
	"time"

	"github.com/gotracker/gomixing/volume"
	device "github.com/gotracker/gosound"
	"github.com/gotracker/voice/render"

	"gotracker/internal/player/feature"
)

// Playback is an interface for rendering a song to output data
type Playback interface {
	Update(time.Duration, chan<- *device.PremixData) error
	Generate(time.Duration) (*device.PremixData, error)

	GetSongData() SongData

	GetNumChannels() int
	GetNumOrders() int
	SetNextOrder(OrderIdx) error
	SetNextRow(RowIdx, ...bool) error
	GetCurrentRow() RowIdx
	GetGlobalVolume() volume.Volume
	SetGlobalVolume(volume.Volume)
	Configure([]feature.Feature)
	GetName() string
	CanOrderLoop() bool
	BreakOrder() error
	SetOnEffect(func(Effect))
	IgnoreUnknownEffect() bool

	SetupSampler(int, int, int) error
	GetSampleRate() float32
	GetOPL2Chip() render.OPL2Chip
}
