package intf

import (
	"time"

	"github.com/gotracker/gomixing/volume"
	device "github.com/gotracker/gosound"

	"gotracker/internal/player/feature"
)

// Playback is an interface for rendering a song to output data
type Playback interface {
	Update(time.Duration, chan<- *device.PremixData) error

	GetSongData() SongData

	GetNumChannels() int
	GetNumOrders() int
	SetNextOrder(OrderIdx)
	SetNextRow(RowIdx)
	GetGlobalVolume() volume.Volume
	SetGlobalVolume(volume.Volume)
	DisableFeatures([]feature.Feature)
	GetName() string
	CanOrderLoop() bool
	BreakOrder()

	SetupSampler(int, int, int) error
}
