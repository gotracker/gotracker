package output

import (
	"github.com/gotracker/voice/period"
	"github.com/gotracker/voice/render"

	"github.com/gotracker/gomixing/volume"
)

type ConfigIntf interface {
	SetupSampler(int, int, int) error
	GetSampleRate() period.Frequency
	GetOPL2Chip() render.OPL2Chip
	GetGlobalVolume() volume.Volume
	SetGlobalVolume(volume.Volume)
}
