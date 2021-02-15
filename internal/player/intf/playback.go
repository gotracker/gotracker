package intf

import (
	"time"

	"github.com/gotracker/gomixing/volume"
	device "github.com/gotracker/gosound"

	"gotracker/internal/player/feature"
	"gotracker/internal/player/render"
)

// Playback is an interface for rendering a song to output data
type Playback interface {
	Update(time.Duration, chan<- *device.PremixData) error
	Generate(time.Duration) (*device.PremixData, error)

	GetSongData() SongData

	GetNumChannels() int
	GetNumOrders() int
	SetNextOrder(OrderIdx)
	SetNextRow(RowIdx, ...bool)
	GetCurrentRow() RowIdx
	GetGlobalVolume() volume.Volume
	SetGlobalVolume(volume.Volume)
	DisableFeatures([]feature.Feature)
	GetName() string
	CanOrderLoop() bool
	BreakOrder()
	SetOnEffect(func(Effect))
	IgnoreUnknownEffect() bool

	SetupSampler(int, int, int) error
	GetSampleRate() float32
	GetOPL2Chip() render.OPL2Chip
}
